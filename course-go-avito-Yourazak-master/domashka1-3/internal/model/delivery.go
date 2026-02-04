package model

import "time"

type Delivery struct {
	ID         int       `json:"id"`
	CourierID  int       `json:"courier_id"`
	OrderID    string    `json:"order_id"`
	AssignedAt time.Time `json:"assigned_at"`
	Deadline   time.Time `json:"deadline"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
