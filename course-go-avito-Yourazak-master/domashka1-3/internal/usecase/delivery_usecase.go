package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"avito-courier/internal/model"
	"avito-courier/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoAvailableCourier   = errors.New("no available courier")
	ErrOrderAlreadyAssigned = errors.New("order already assigned")
)

type IDeliveryUsecase interface {
	Assign(ctx context.Context, orderID string) (model.Delivery, model.Courier, error)
	Unassign(ctx context.Context, orderID string) error
	AutoRelease(ctx context.Context, interval time.Duration)
	AssignForEvent(ctx context.Context, orderID string) error
	UnassignForEvent(ctx context.Context, orderID string) error
	CompleteForEvent(ctx context.Context, orderID string) error
	GetByOrderID(ctx context.Context, orderID string) (model.Delivery, error)
	Create(ctx context.Context, d *model.Delivery) error
	DeleteByOrderID(ctx context.Context, orderID string) error
}

type DeliveryUsecase struct {
	pool         *pgxpool.Pool
	courierRepo  repository.CourierRepository
	deliveryRepo repository.DeliveryRepository
	factory      *DeliveryTimeFactory
}

func NewDeliveryUsecase(pool *pgxpool.Pool, cr repository.CourierRepository, dr repository.DeliveryRepository, f *DeliveryTimeFactory) *DeliveryUsecase {
	return &DeliveryUsecase{
		pool:         pool,
		courierRepo:  cr,
		deliveryRepo: dr,
		factory:      f,
	}
}

func (u *DeliveryUsecase) Assign(ctx context.Context, orderID string) (model.Delivery, model.Courier, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Delivery{}, model.Courier{}, err
	}
	defer tx.Rollback(ctx)

	delivery, err := u.deliveryRepo.GetByOrderIDTx(ctx, tx, orderID)
	if err == nil && delivery.ID > 0 {
		return model.Delivery{}, model.Courier{}, ErrOrderAlreadyAssigned
	}

	courier, err := u.courierRepo.FindAvailableCourier(ctx)
	if err != nil {
		if err == repository.ErrNotFound {
			return model.Delivery{}, model.Courier{}, ErrNoAvailableCourier
		}
		return model.Delivery{}, model.Courier{}, err
	}

	now := time.Now().UTC()
	deadline := u.factory.Deadline(now, courier.TransportType)

	newDelivery := &model.Delivery{
		CourierID:  courier.ID,
		OrderID:    orderID,
		AssignedAt: now,
		Deadline:   deadline,
	}

	if err := u.deliveryRepo.CreateTx(ctx, tx, newDelivery); err != nil {
		return model.Delivery{}, model.Courier{}, err
	}

	if err := u.courierRepo.UpdateStatusTx(ctx, tx, courier.ID, "busy"); err != nil {
		return model.Delivery{}, model.Courier{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Delivery{}, model.Courier{}, err
	}

	return *newDelivery, courier, nil
}

func (u *DeliveryUsecase) Unassign(ctx context.Context, orderID string) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	courierID, err := u.deliveryRepo.DeleteByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		if err.Error() == "delivery not found" {
			return errors.New("delivery not found")
		}
		return err
	}

	if err := u.courierRepo.UpdateStatusTx(ctx, tx, courierID, "available"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *DeliveryUsecase) AssignForEvent(ctx context.Context, orderID string) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	exists, err := u.deliveryRepo.CheckOrderExistsTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Order %s already assigned, skipping", orderID)
		return nil
	}

	courier, err := u.courierRepo.FindAvailableCourier(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	deadline := u.factory.Deadline(now, courier.TransportType)

	delivery := &model.Delivery{
		CourierID:  courier.ID,
		OrderID:    orderID,
		AssignedAt: now,
		Deadline:   deadline,
	}

	if err := u.deliveryRepo.CreateTx(ctx, tx, delivery); err != nil {
		return err
	}

	if err := u.courierRepo.UpdateStatusTx(ctx, tx, courier.ID, "busy"); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (u *DeliveryUsecase) UnassignForEvent(ctx context.Context, orderID string) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	delivery, err := u.deliveryRepo.GetByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		if err.Error() == "delivery not found" {
			log.Printf("Delivery for order %s not found, skipping", orderID)
			return nil
		}
		return err
	}

	if _, err := u.deliveryRepo.DeleteByOrderIDTx(ctx, tx, orderID); err != nil {
		return err
	}

	if err := u.courierRepo.UpdateStatusTx(ctx, tx, delivery.CourierID, "available"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *DeliveryUsecase) CompleteForEvent(ctx context.Context, orderID string) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	delivery, err := u.deliveryRepo.GetByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		if err.Error() == "delivery not found" {
			log.Printf("Delivery for order %s not found, skipping completion", orderID)
			return nil
		}
		return err
	}

	if err := u.courierRepo.UpdateStatusTx(ctx, tx, delivery.CourierID, "available"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *DeliveryUsecase) GetByOrderID(ctx context.Context, orderID string) (model.Delivery, error) {
	return u.deliveryRepo.GetByOrderID(ctx, orderID)
}

func (u *DeliveryUsecase) Create(ctx context.Context, d *model.Delivery) error {
	return u.deliveryRepo.Create(ctx, d)
}

func (u *DeliveryUsecase) DeleteByOrderID(ctx context.Context, orderID string) error {
	return u.deliveryRepo.DeleteByOrderID(ctx, orderID)
}

func (u *DeliveryUsecase) AutoRelease(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.pool.Exec(ctx, `
				UPDATE couriers
				SET status = 'available', updated_at = NOW()
				FROM (
					SELECT DISTINCT courier_id
					FROM deliveries
					WHERE deadline < NOW()
				) d
				WHERE couriers.id = d.courier_id
				  AND couriers.status = 'busy';
			`)
		}
	}
}
