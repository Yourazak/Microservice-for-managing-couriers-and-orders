package handler

import (
	"avito-courier/internal/usecase"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito-courier/internal/model"
	_ "avito-courier/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCourierService struct {
	mock.Mock
}

func (m *MockCourierService) GetByID(ctx context.Context, id int) (model.Courier, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Courier), args.Error(1)
}

func (m *MockCourierService) GetAll(ctx context.Context) ([]model.Courier, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Courier), args.Error(1)
}

func (m *MockCourierService) Create(ctx context.Context, c *model.Courier) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCourierService) Update(ctx context.Context, c *model.Courier) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func TestCourierHandler_GetAll_Success(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	expectedCouriers := []model.Courier{
		{ID: 1, Name: "John", Phone: "+79123456789", Status: "available", TransportType: "car"},
		{ID: 2, Name: "Jane", Phone: "+79123456780", Status: "busy", TransportType: "bike"},
	}

	mockService.On("GetAll", mock.Anything).Return(expectedCouriers, nil)

	req := httptest.NewRequest("GET", "/couriers", nil)
	rr := httptest.NewRecorder()

	handler.GetAll(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []model.CourierDTO
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "John", response[0].Name)
}

func TestCourierHandler_GetAll_MethodNotAllowed(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	req := httptest.NewRequest("POST", "/couriers", nil)
	rr := httptest.NewRecorder()

	handler.GetAll(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestCourierHandler_GetByID_Success(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	expectedCourier := model.Courier{
		ID: 1, Name: "John", Phone: "+79123456789", Status: "available",
		TransportType: "car", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	mockService.On("GetByID", mock.Anything, 1).Return(expectedCourier, nil)

	req := httptest.NewRequest("GET", "/courier/1", nil)
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.CourierDTO
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, "John", response.Name)
}

func TestCourierHandler_GetByID_NotFound(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	mockService.On("GetByID", mock.Anything, 999).Return(model.Courier{}, usecase.ErrNotFound)

	req := httptest.NewRequest("GET", "/courier/999", nil)
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "курьер не найден")
}

func TestCourierHandler_GetByID_InvalidID(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	req := httptest.NewRequest("GET", "/courier/invalid", nil)
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "id невалидный")
}

func TestCourierHandler_Create_Success(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	courierDTO := model.CourierDTO{
		Name:          "New Courier",
		Phone:         "+79123456789",
		Status:        "available",
		TransportType: "car",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*model.Courier")).Return(nil)

	body, _ := json.Marshal(courierDTO)
	req := httptest.NewRequest("POST", "/couriers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*model.Courier"))
}

func TestCourierHandler_Create_Conflict(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	courierDTO := model.CourierDTO{
		Name:          "New Courier",
		Phone:         "+79123456789",
		Status:        "available",
		TransportType: "car",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*model.Courier")).Return(usecase.ErrConflict)

	body, _ := json.Marshal(courierDTO)
	req := httptest.NewRequest("POST", "/couriers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "курьер с таким телефоном уже существует")
}

func TestCourierHandler_Create_InvalidJSON(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	req := httptest.NewRequest("POST", "/couriers", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCourierHandler_Update_Success(t *testing.T) {
	mockService := new(MockCourierService)
	handler := NewCourierHandler(mockService)

	courierDTO := model.CourierDTO{
		ID: 1, Name: "Updated Courier", Phone: "+79123456789", Status: "available",
	}

	mockService.On("Update", mock.Anything, mock.AnythingOfType("*model.Courier")).Return(nil)

	body, _ := json.Marshal(courierDTO)
	req := httptest.NewRequest("PUT", "/couriers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*model.Courier"))
}
