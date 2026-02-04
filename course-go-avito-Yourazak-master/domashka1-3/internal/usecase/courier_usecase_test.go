package usecase

import (
	"context"
	"testing"

	"avito-courier/internal/model"
	"avito-courier/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCourierRepository struct {
	mock.Mock
}

func (m *MockCourierRepository) GetByID(ctx context.Context, id int) (model.Courier, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Courier), args.Error(1)
}

func (m *MockCourierRepository) GetAll(ctx context.Context) ([]model.Courier, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Courier), args.Error(1)
}

func (m *MockCourierRepository) FindAvailableCourier(ctx context.Context) (model.Courier, error) {
	args := m.Called(ctx)
	return args.Get(0).(model.Courier), args.Error(1)
}

func (m *MockCourierRepository) Create(ctx context.Context, c *model.Courier) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCourierRepository) Update(ctx context.Context, c *model.Courier) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCourierRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestCourierService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	expectedCourier := model.Courier{
		ID: 1, Name: "John", Phone: "+79123456789", Status: "available",
	}

	mockRepo.On("GetByID", mock.Anything, 1).Return(expectedCourier, nil)

	courier, err := service.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedCourier, courier)
	mockRepo.AssertCalled(t, "GetByID", mock.Anything, 1)
}

func TestCourierService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	mockRepo.On("GetByID", mock.Anything, 999).Return(model.Courier{}, repository.ErrNotFound)

	courier, err := service.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Equal(t, model.Courier{}, courier)
}

func TestCourierService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	courier, err := service.GetByID(context.Background(), 0)

	assert.Error(t, err)
	assert.Equal(t, ErrBadInput, err)
	assert.Equal(t, model.Courier{}, courier)
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestCourierService_GetAll_Success(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	expectedCouriers := []model.Courier{
		{ID: 1, Name: "John", Phone: "+79123456789", Status: "available"},
		{ID: 2, Name: "Jane", Phone: "+79123456780", Status: "busy"},
	}

	mockRepo.On("GetAll", mock.Anything).Return(expectedCouriers, nil)

	couriers, err := service.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedCouriers, couriers)
	mockRepo.AssertCalled(t, "GetAll", mock.Anything)
}

func TestCourierService_Create_Success(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	courier := &model.Courier{
		Name:          "New Courier",
		Phone:         "+79123456789",
		Status:        "available",
		TransportType: "car",
	}

	mockRepo.On("Create", mock.Anything, courier).Return(nil)

	err := service.Create(context.Background(), courier)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Create", mock.Anything, courier)
}

func TestCourierService_Create_InvalidData(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	testCases := []struct {
		name    string
		courier *model.Courier
	}{
		{"Empty Name", &model.Courier{Name: "", Phone: "+79123456789", Status: "available"}},
		{"Empty Phone", &model.Courier{Name: "John", Phone: "", Status: "available"}},
		{"Empty Status", &model.Courier{Name: "John", Phone: "+79123456789", Status: ""}},
		{"Invalid Status", &model.Courier{Name: "John", Phone: "+79123456789", Status: "invalid"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.Create(context.Background(), tc.courier)
			assert.Error(t, err)
			assert.Equal(t, ErrBadInput, err)
		})
	}
}

func TestCourierService_Update_Success(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	courier := &model.Courier{
		ID: 1, Name: "Updated Courier", Phone: "+79123456789", Status: "available",
	}

	mockRepo.On("Update", mock.Anything, courier).Return(nil)

	err := service.Update(context.Background(), courier)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "Update", mock.Anything, courier)
}

func TestCourierService_Update_InvalidID(t *testing.T) {
	mockRepo := new(MockCourierRepository)
	service := NewCourierUsecase(mockRepo)

	courier := &model.Courier{
		ID: 0, Name: "Updated Courier", Phone: "+79123456789", Status: "available",
	}

	err := service.Update(context.Background(), courier)

	assert.Error(t, err)
	assert.Equal(t, ErrBadInput, err)
	mockRepo.AssertNotCalled(t, "Update")
}
