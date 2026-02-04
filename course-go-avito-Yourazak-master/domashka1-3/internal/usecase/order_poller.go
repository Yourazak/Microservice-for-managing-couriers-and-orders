package usecase

import (
	"context"
	"log"
	"time"

	"avito-courier/internal/gateway/order"
)

type OrderPoller struct {
	gateway    order.OrderGateway
	deliveryUC *DeliveryUsecase
	interval   time.Duration
	lastFetch  time.Time
}

func NewOrderPoller(gateway order.OrderGateway, deliveryUC *DeliveryUsecase) *OrderPoller {
	return &OrderPoller{
		gateway:    gateway,
		deliveryUC: deliveryUC,
		interval:   5 * time.Second,
		lastFetch:  time.Now().Add(-5 * time.Second),
	}
}

func (p *OrderPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	log.Printf("Order poller started (ticker: %v)", p.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Order poller stopped")
			return
		case t := <-ticker.C:
			p.processTick(ctx, t)
		}
	}
}

func (p *OrderPoller) processTick(ctx context.Context, tickTime time.Time) {
	start := time.Now()
	cursor := tickTime.Add(-5 * time.Second)

	orders, err := p.gateway.GetOrdersByCursor(ctx, cursor)
	if err != nil {
		log.Printf("Failed to fetch orders: %v", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, o := range orders {
		if o.Status == "created" {
			if err := p.deliveryUC.AssignForEvent(ctx, o.OrderID); err != nil {
				log.Printf("Failed to assign courier to order %s: %v", o.OrderID, err)
			}
		}
	}

	p.lastFetch = tickTime
	log.Printf("Tick processed in %v (orders: %d)", time.Since(start), len(orders))
}
