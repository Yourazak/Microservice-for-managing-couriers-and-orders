package model

import "time"

type ExternalOrder struct {
	ID        string    `json:"id"`
	Weight    float64   `json:"weight"`
	Region    int       `json:"region"`
	Cost      int       `json:"cost"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderEvent struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (e *OrderEvent) Validate() bool {
	if e.OrderID == "" || e.Status == "" || e.CreatedAt.IsZero() {
		return false
	}

	validStatuses := map[string]bool{
		"created":   true,
		"cancelled": true,
		"completed": true,
	}

	return validStatuses[e.Status]
}
