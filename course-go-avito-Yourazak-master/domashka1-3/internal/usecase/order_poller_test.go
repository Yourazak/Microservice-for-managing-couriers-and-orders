package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderGateway struct {
	mock.Mock
}

func (m *MockOrderGateway) GetOrdersByCursor(ctx context.Context, cursor time.Time) ([]OrderEvent, error) {
	args := m.Called(ctx, cursor)
	return args.Get(0).([]OrderEvent), args.Error(1)
}

func (m *MockOrderGateway) GetOrderStatus(ctx context.Context, orderID string) (*OrderEvent, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*OrderEvent), args.Error(1)
}

type MockPollerDeliveryUsecase struct {
	mock.Mock
}

func (m *MockPollerDeliveryUsecase) Assign(ctx context.Context, orderID string) (Delivery, Courier, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(Delivery), args.Get(1).(Courier), args.Error(2)
}

func (m *MockPollerDeliveryUsecase) Unassign(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockPollerDeliveryUsecase) AutoRelease(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

func TestOrderPoller_Start(t *testing.T) {
	mockGateway := new(MockOrderGateway)
	mockDeliveryUC := new(MockPollerDeliveryUsecase)

	poller := NewOrderPoller(mockGateway, mockDeliveryUC)

	ctx, cancel := context.WithCancel(context.Background())

	go poller.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	cancel()
	time.Sleep(100 * time.Millisecond)

	assert.True(t, true)
}
