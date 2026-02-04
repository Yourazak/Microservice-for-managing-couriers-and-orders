package usecase

import (
	"context"
	"strings"

	"avito-courier/internal/model"
	"avito-courier/internal/repository"
)

var (
	ErrNotFound = repository.ErrNotFound
	ErrConflict = repository.ErrConflict
	ErrBadInput = repository.ErrBadInput
)

type CourierUsecase interface {
	GetByID(ctx context.Context, id int) (model.Courier, error)
	GetAll(ctx context.Context) ([]model.Courier, error)
	Create(ctx context.Context, c *model.Courier) error
	Update(ctx context.Context, c *model.Courier) error
}

type courierUsecase struct {
	repo       repository.CourierRepository
	deliveryUC *DeliveryUsecase
}

func NewCourierUsecase(r repository.CourierRepository, deliveryUC *DeliveryUsecase) CourierUsecase {
	return &courierUsecase{
		repo:       r,
		deliveryUC: deliveryUC,
	}
}

var validStatus = map[string]bool{
	"available": true,
	"busy":      true,
	"paused":    true,
}

func (u *courierUsecase) GetByID(ctx context.Context, id int) (model.Courier, error) {
	if id <= 0 {
		return model.Courier{}, ErrBadInput
	}
	c, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return model.Courier{}, err
	}
	return c, nil
}

func (u *courierUsecase) GetAll(ctx context.Context) ([]model.Courier, error) {
	return u.repo.GetAll(ctx)
}

func (u *courierUsecase) Create(ctx context.Context, c *model.Courier) error {
	if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Phone) == "" || strings.TrimSpace(c.Status) == "" {
		return ErrBadInput
	}
	if !validStatus[c.Status] {
		return ErrBadInput
	}
	if err := u.repo.Create(ctx, c); err != nil {
		return err
	}
	return nil
}

func (u *courierUsecase) Update(ctx context.Context, c *model.Courier) error {
	if c.ID <= 0 {
		return ErrBadInput
	}

	if strings.TrimSpace(c.Phone) == "" || strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Status) == "" {
		return ErrBadInput
	}
	if !validStatus[c.Status] {
		return ErrBadInput
	}
	if err := u.repo.Update(ctx, c); err != nil {
		return err
	}
	return nil
}
