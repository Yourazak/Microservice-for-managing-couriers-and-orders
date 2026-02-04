package model

import "time"

type CourierDTO struct {
	ID            int       `json:"id,omitempty"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone"`
	Status        string    `json:"status"`
	TransportType string    `json:"transport_type"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

func ToDTO(c Courier) CourierDTO {
	return CourierDTO{
		ID:            c.ID,
		Name:          c.Name,
		Phone:         c.Phone,
		Status:        c.Status,
		TransportType: c.TransportType,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func FromDTO(d CourierDTO) Courier {
	return Courier{
		ID:            d.ID,
		Name:          d.Name,
		Phone:         d.Phone,
		Status:        d.Status,
		TransportType: d.TransportType,
	}
}
