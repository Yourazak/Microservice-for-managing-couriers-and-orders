package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToDTO(t *testing.T) {
	now := time.Now()
	courier := Courier{
		ID:            1,
		Name:          "John",
		Phone:         "+79123456789",
		Status:        "available",
		TransportType: "car",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	dto := ToDTO(courier)

	assert.Equal(t, courier.ID, dto.ID)
	assert.Equal(t, courier.Name, dto.Name)
	assert.Equal(t, courier.Phone, dto.Phone)
	assert.Equal(t, courier.Status, dto.Status)
	assert.Equal(t, courier.TransportType, dto.TransportType)
	assert.Equal(t, courier.CreatedAt, dto.CreatedAt)
	assert.Equal(t, courier.UpdatedAt, dto.UpdatedAt)
}

func TestFromDTO(t *testing.T) {
	now := time.Now()
	dto := CourierDTO{
		ID:            1,
		Name:          "John",
		Phone:         "+79123456789",
		Status:        "available",
		TransportType: "car",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	courier := FromDTO(dto)

	assert.Equal(t, dto.ID, courier.ID)
	assert.Equal(t, dto.Name, courier.Name)
	assert.Equal(t, dto.Phone, courier.Phone)
	assert.Equal(t, dto.Status, courier.Status)
	assert.Equal(t, dto.TransportType, courier.TransportType)
}
