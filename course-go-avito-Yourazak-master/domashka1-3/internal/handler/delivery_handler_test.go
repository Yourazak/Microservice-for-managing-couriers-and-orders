package handler

import (
	"avito-courier/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito-courier/internal/model"
	"avito-courier/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDeliveryUsecase struct {
	mock.Mock
}

func (m *MockDeliveryUsecase) Assign(ctx context.Context, orderID string) (model.Delivery, model.Courier, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(model.Delivery), args.Get(1).(model.Courier), args.Error(2)
}

func (m *MockDeliveryUsecase) Unassign(ctx context.Context, orderID string) (int, error) {
	args := m.Called(ctx, orderID)
	return args.Int(0), args.Error(1)
}

func (m *MockDeliveryUsecase) AutoRelease(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

func TestDeliveryHandler_Assign_Success(t *testing.T) {
	mockUsecase := new(MockDeliveryUsecase)
	handler := NewDeliveryHandler(mockUsecase)

	now := time.Now().UTC()
	deadline := now.Add(2 * time.Hour)

	expectedDelivery := model.Delivery{
		OrderID:    "order-123",
		CourierID:  1,
		AssignedAt: now,
		Deadline:   deadline,
	}

	expectedCourier := model.Courier{
		ID:            1,
		Name:          "John",
		TransportType: "car",
	}

	mockUsecase.On("Assign", mock.Anything, "order-123").Return(expectedDelivery, expectedCourier, nil)

	reqBody := assignReq{OrderID: "order-123"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/deliveries/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Assign(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response assignResp
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.CourierID)
	assert.Equal(t, "order-123", response.OrderID)
	assert.Equal(t, "car", response.TransportType)
}

func TestDeliveryHandler_Assign_NoAvailableCourier(t *testing.T) {
	mockUsecase := new(MockDeliveryUsecase)
	handler := NewDeliveryHandler(mockUsecase)

	mockUsecase.On("Assign", mock.Anything, "order-123").Return(model.Delivery{}, model.Courier{}, usecase.ErrNoAvailableCourier)

	reqBody := assignReq{OrderID: "order-123"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/deliveries/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Assign(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "no available couriers")
}

func TestDeliveryHandler_Assign_BadRequest(t *testing.T) {
	mockUsecase := new(MockDeliveryUsecase)
	handler := NewDeliveryHandler(mockUsecase)

	reqBody := assignReq{OrderID: ""}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/deliveries/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Assign(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeliveryHandler_Unassign_Success(t *testing.T) {
	mockUsecase := new(MockDeliveryUsecase)
	handler := NewDeliveryHandler(mockUsecase)

	mockUsecase.On("Unassign", mock.Anything, "order-123").Return(1, nil)

	reqBody := unassignReq{OrderID: "order-123"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/deliveries/unassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Unassign(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response unassignResp
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "order-123", response.OrderID)
	assert.Equal(t, "unassigned", response.Status)
	assert.Equal(t, 1, response.CourierID)
}

func TestDeliveryHandler_Unassign_NotFound(t *testing.T) {
	mockUsecase := new(MockDeliveryUsecase)
	handler := NewDeliveryHandler(mockUsecase)

	mockUsecase.On("Unassign", mock.Anything, "order-999").Return(0, repository.ErrDeliveryNotFound)

	reqBody := unassignReq{OrderID: "order-999"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/deliveries/unassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Unassign(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
