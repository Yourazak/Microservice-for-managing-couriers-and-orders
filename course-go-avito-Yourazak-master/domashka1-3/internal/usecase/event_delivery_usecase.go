package usecase

import (
	"context"

	"avito-courier/internal/model"
)

type EventDeliveryUsecase struct {
	deliveryUC *DeliveryUsecase
}

func NewEventDeliveryUsecase(deliveryUC *DeliveryUsecase) *EventDeliveryUsecase {
	return &EventDeliveryUsecase{
		deliveryUC: deliveryUC,
	}
}

func (uc *EventDeliveryUsecase) HandleCreated(ctx context.Context, event model.OrderEvent) error {
	return uc.deliveryUC.AssignForEvent(ctx, event.OrderID)
}

func (uc *EventDeliveryUsecase) HandleCancelled(ctx context.Context, event model.OrderEvent) error {
	return uc.deliveryUC.UnassignForEvent(ctx, event.OrderID)
}

func (uc *EventDeliveryUsecase) HandleCompleted(ctx context.Context, event model.OrderEvent) error {
	return uc.deliveryUC.CompleteForEvent(ctx, event.OrderID)
}
