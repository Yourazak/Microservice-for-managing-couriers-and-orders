package repository

import (
	"context"
	"errors"

	"avito-courier/internal/model"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrBadInput         = errors.New("bad input")
	ErrDeliveryNotFound = errors.New("delivery not found")
)

type CourierRepository interface {
	GetByID(ctx context.Context, id int) (model.Courier, error)
	GetAll(ctx context.Context) ([]model.Courier, error)
	Create(ctx context.Context, c *model.Courier) error
	Update(ctx context.Context, c *model.Courier) error
	FindAvailableCourier(ctx context.Context) (model.Courier, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateStatusTx(ctx context.Context, tx pgx.Tx, id int, status string) error
}

type courierRepo struct {
	pool *pgxpool.Pool
}

func NewCourierRepository(pool *pgxpool.Pool) CourierRepository {
	return &courierRepo{pool: pool}
}

func (r *courierRepo) GetByID(ctx context.Context, id int) (model.Courier, error) {
	var c model.Courier
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, phone, status, transport_type, created_at, updated_at FROM couriers WHERE id=$1`, id).
		Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Courier{}, ErrNotFound
		}
		return model.Courier{}, err
	}
	return c, nil
}

func (r *courierRepo) GetAll(ctx context.Context) ([]model.Courier, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, phone, status, transport_type, created_at, updated_at FROM couriers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Courier
	for rows.Next() {
		var c model.Courier
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *courierRepo) Create(ctx context.Context, c *model.Courier) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO couriers (name, phone, status, transport_type, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, NOW(), NOW()) 
		 RETURNING id, created_at, updated_at`,
		c.Name, c.Phone, c.Status, c.TransportType)
	if err := row.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrConflict
			}
		}
		return err
	}
	return nil
}

func (r *courierRepo) Update(ctx context.Context, c *model.Courier) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE couriers 
		 SET name = $1, phone = $2, status = $3, transport_type = $4, updated_at = NOW() 
		 WHERE id = $5`,
		c.Name, c.Phone, c.Status, c.TransportType, c.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrConflict
			}
		}
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *courierRepo) FindAvailableCourier(ctx context.Context) (model.Courier, error) {
	var c model.Courier
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, phone, status, transport_type, created_at, updated_at
		FROM couriers
		WHERE status = 'available'
		ORDER BY created_at ASC
		LIMIT 1
	`).Scan(&c.ID, &c.Name, &c.Phone, &c.Status, &c.TransportType, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Courier{}, ErrNotFound
		}
		return model.Courier{}, err
	}
	return c, nil
}

func (r *courierRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE couriers
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, status, id)
	return err
}

func (r *courierRepo) UpdateStatusTx(ctx context.Context, tx pgx.Tx, id int, status string) error {
	_, err := tx.Exec(ctx, `
		UPDATE couriers
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, status, id)
	return err
}
