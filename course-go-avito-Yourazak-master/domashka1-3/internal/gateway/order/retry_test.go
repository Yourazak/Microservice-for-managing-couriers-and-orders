package order

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito-courier/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestRetryGateway(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"order_id":"test","status":"created","created_at":"2025-01-01T00:00:00Z"}]`))
	}))
	defer server.Close()

	cfg := &config.Config{
		ServiceOrderURL: server.URL,
	}

	gateway := NewHTTPOrderGatewayWithRetry(cfg)

	ctx := context.Background()
	cursor := time.Now()

	orders, err := gateway.GetOrdersByCursor(ctx, cursor)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(orders))
	assert.Equal(t, "test", orders[0].OrderID)
	assert.Equal(t, 3, attempts)
}

func TestRetryExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{
		ServiceOrderURL: server.URL,
	}

	gateway := NewHTTPOrderGatewayWithRetry(cfg)
	gateway.retryConfig.MaxRetries = 2

	ctx := context.Background()
	cursor := time.Now()

	_, err := gateway.GetOrdersByCursor(ctx, cursor)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 2 retries")
	assert.Equal(t, 3, attempts)
}
