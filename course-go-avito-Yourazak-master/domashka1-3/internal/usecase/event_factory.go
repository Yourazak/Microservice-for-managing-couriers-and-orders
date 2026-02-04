package usecase

import (
	"avito-courier/internal/model"
	"context"
)

type EventHandler interface {
	Handle(ctx context.Context, event model.OrderEvent) error
}

type EventHandlerFactory struct {
	eventDeliveryUC *EventDeliveryUsecase
}

func NewEventHandlerFactory(eventDeliveryUC *EventDeliveryUsecase) *EventHandlerFactory {
	return &EventHandlerFactory{
		eventDeliveryUC: eventDeliveryUC,
	}
}

func (f *EventHandlerFactory) GetHandler(status string) EventHandler {
	switch status {
	case "created":
		return &CreatedHandler{f.eventDeliveryUC}
	case "cancelled":
		return &CancelledHandler{f.eventDeliveryUC}
	case "completed":
		return &CompletedHandler{f.eventDeliveryUC}
	default:
		return nil
	}
}

type CreatedHandler struct {
	eventDeliveryUC *EventDeliveryUsecase
}

func (h *CreatedHandler) Handle(ctx context.Context, event model.OrderEvent) error {
	return h.eventDeliveryUC.HandleCreated(ctx, event)
}

type CancelledHandler struct {
	eventDeliveryUC *EventDeliveryUsecase
}

func (h *CancelledHandler) Handle(ctx context.Context, event model.OrderEvent) error {
	return h.eventDeliveryUC.HandleCancelled(ctx, event)
}

type CompletedHandler struct {
	eventDeliveryUC *EventDeliveryUsecase
}

func (h *CompletedHandler) Handle(ctx context.Context, event model.OrderEvent) error {
	return h.eventDeliveryUC.HandleCompleted(ctx, event)
}
