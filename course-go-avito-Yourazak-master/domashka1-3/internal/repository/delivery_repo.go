package repository

import (
	"context"
	"errors"
	"time"

	"avito-courier/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryRepository interface {
	Create(ctx context.Context, d *model.Delivery) error
	CreateTx(ctx context.Context, tx pgx.Tx, d *model.Delivery) error
	GetByOrderID(ctx context.Context, orderID string) (model.Delivery, error)
	GetByOrderIDTx(ctx context.Context, tx pgx.Tx, orderID string) (model.Delivery, error)
	UpdateStatus(ctx context.Context, orderID, status string) error
	DeleteByOrderID(ctx context.Context, orderID string) error
	ReleaseExpired(ctx context.Context, before time.Time) ([]string, error)

	CheckOrderExistsTx(ctx context.Context, tx pgx.Tx, orderID string) (bool, error)
	DeleteByOrderIDTx(ctx context.Context, tx pgx.Tx, orderID string) (int, error)
}

type deliveryRepo struct {
	pool *pgxpool.Pool
}

func NewDeliveryRepository(pool *pgxpool.Pool) DeliveryRepository {
	return &deliveryRepo{pool: pool}
}

func (r *deliveryRepo) Create(ctx context.Context, d *model.Delivery) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := r.CreateTx(ctx, tx, d); err != nil {
		return err
	}

	if err := r.updateCourierStatusTx(ctx, tx, d.CourierID, "busy"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *deliveryRepo) CreateTx(ctx context.Context, tx pgx.Tx, d *model.Delivery) error {
	return tx.QueryRow(ctx,
		`INSERT INTO deliveries (order_id, courier_id, status, created_at, updated_at, assigned_at)
		 VALUES ($1, $2, $3, NOW(), NOW(), NOW())
		 RETURNING id, created_at, updated_at, assigned_at`,
		d.OrderID, d.CourierID, "assigned").
		Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt, &d.AssignedAt)
}

func (r *deliveryRepo) updateCourierStatusTx(ctx context.Context, tx pgx.Tx, courierID int, status string) error {
	_, err := tx.Exec(ctx,
		`UPDATE couriers SET status=$1, updated_at=NOW() WHERE id=$2`,
		status, courierID)
	return err
}

func (r *deliveryRepo) GetByOrderID(ctx context.Context, orderID string) (model.Delivery, error) {
	var d model.Delivery
	err := r.pool.QueryRow(ctx,
		`SELECT id, order_id, courier_id, status, created_at, updated_at, assigned_at
		 FROM deliveries WHERE order_id=$1`,
		orderID).
		Scan(&d.ID, &d.OrderID, &d.CourierID, &d.Status, &d.CreatedAt, &d.UpdatedAt, &d.AssignedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Delivery{}, errors.New("delivery not found")
		}
		return model.Delivery{}, err
	}
	return d, nil
}

func (r *deliveryRepo) GetByOrderIDTx(ctx context.Context, tx pgx.Tx, orderID string) (model.Delivery, error) {
	var d model.Delivery
	err := tx.QueryRow(ctx,
		`SELECT id, order_id, courier_id, status, created_at, updated_at, assigned_at
		 FROM deliveries WHERE order_id=$1`,
		orderID).
		Scan(&d.ID, &d.OrderID, &d.CourierID, &d.Status, &d.CreatedAt, &d.UpdatedAt, &d.AssignedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Delivery{}, errors.New("delivery not found")
		}
		return model.Delivery{}, err
	}
	return d, nil
}

func (r *deliveryRepo) UpdateStatus(ctx context.Context, orderID, status string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE deliveries SET status=$1, updated_at=NOW() WHERE order_id=$2`,
		status, orderID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("delivery not found")
	}
	return nil
}

func (r *deliveryRepo) DeleteByOrderID(ctx context.Context, orderID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var courierID int
	err = tx.QueryRow(ctx,
		`SELECT courier_id FROM deliveries WHERE order_id=$1`,
		orderID).Scan(&courierID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("delivery not found")
		}
		return err
	}

	result, err := tx.Exec(ctx,
		`DELETE FROM deliveries WHERE order_id=$1`,
		orderID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("delivery not found")
	}

	if err := r.updateCourierStatusTx(ctx, tx, courierID, "available"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *deliveryRepo) ReleaseExpired(ctx context.Context, before time.Time) ([]string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`SELECT order_id, courier_id FROM deliveries 
		 WHERE status='assigned' AND assigned_at < $1`,
		before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderIDs []string
	var courierIDs []int
	for rows.Next() {
		var orderID string
		var courierID int
		if err := rows.Scan(&orderID, &courierID); err != nil {
			return nil, err
		}
		orderIDs = append(orderIDs, orderID)
		courierIDs = append(courierIDs, courierID)
	}

	if len(orderIDs) == 0 {
		return nil, tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx,
		`UPDATE deliveries SET status='expired', updated_at=NOW() 
		 WHERE status='assigned' AND assigned_at < $1`,
		before)
	if err != nil {
		return nil, err
	}

	for _, courierID := range courierIDs {
		if err := r.updateCourierStatusTx(ctx, tx, courierID, "available"); err != nil {
			return nil, err
		}
	}

	return orderIDs, tx.Commit(ctx)
}

func (r *deliveryRepo) CheckOrderExistsTx(ctx context.Context, tx pgx.Tx, orderID string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM deliveries WHERE order_id = $1)`,
		orderID).Scan(&exists)
	return exists, err
}

func (r *deliveryRepo) DeleteByOrderIDTx(ctx context.Context, tx pgx.Tx, orderID string) (int, error) {
	var courierID int

	err := tx.QueryRow(ctx,
		`SELECT courier_id FROM deliveries WHERE order_id = $1`,
		orderID).Scan(&courierID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("delivery not found")
		}
		return 0, err
	}

	result, err := tx.Exec(ctx,
		`DELETE FROM deliveries WHERE order_id = $1`,
		orderID)
	if err != nil {
		return 0, err
	}

	if result.RowsAffected() == 0 {
		return 0, errors.New("delivery not found")
	}

	return courierID, nil
}
