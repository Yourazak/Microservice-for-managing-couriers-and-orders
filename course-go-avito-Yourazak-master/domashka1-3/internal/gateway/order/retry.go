package order

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"avito-courier/internal/config"
	"avito-courier/internal/middleware"
	"avito-courier/internal/model"
)

type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	Jitter          bool
	RetryableStatus []int
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
		RetryableStatus: []int{429, 500, 502, 503, 504},
	}
}

type HTTPOrderGatewayWithRetry struct {
	*HTTPOrderGateway
	retryConfig RetryConfig
}

func NewHTTPOrderGatewayWithRetry(cfg *config.Config) *HTTPOrderGatewayWithRetry {
	gateway := NewHTTPOrderGateway(cfg)
	return &HTTPOrderGatewayWithRetry{
		HTTPOrderGateway: gateway,
		retryConfig:      DefaultRetryConfig(),
	}
}

func (g *HTTPOrderGatewayWithRetry) GetOrdersByCursor(ctx context.Context, cursor time.Time) ([]model.OrderEvent, error) {
	var lastErr error

	for attempt := 0; attempt <= g.retryConfig.MaxRetries; attempt++ {
		orders, err := g.HTTPOrderGateway.GetOrdersByCursor(ctx, cursor)

		if err == nil {
			return orders, nil
		}

		lastErr = err

		if !g.shouldRetry(err) || attempt == g.retryConfig.MaxRetries {
			break
		}

		middleware.GatewayRetriesTotal.WithLabelValues(
			"GetOrdersByCursor",
			getErrorCode(err),
			fmt.Sprintf("%d", attempt),
		).Inc()

		delay := g.calculateDelay(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", g.retryConfig.MaxRetries, lastErr)
}

func (g *HTTPOrderGatewayWithRetry) GetOrderStatus(ctx context.Context, orderID string) (*model.OrderEvent, error) {
	var lastErr error

	for attempt := 0; attempt <= g.retryConfig.MaxRetries; attempt++ {
		order, err := g.HTTPOrderGateway.GetOrderStatus(ctx, orderID)

		if err == nil {
			return order, nil
		}

		lastErr = err

		if !g.shouldRetry(err) || attempt == g.retryConfig.MaxRetries {
			break
		}

		middleware.GatewayRetriesTotal.WithLabelValues(
			"GetOrderStatus",
			getErrorCode(err),
			fmt.Sprintf("%d", attempt),
		).Inc()

		delay := g.calculateDelay(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", g.retryConfig.MaxRetries, lastErr)
}

func (g *HTTPOrderGatewayWithRetry) shouldRetry(err error) bool {
	errStr := err.Error()

	retryablePhrases := []string{
		"429",
		"Too Many Requests",
		"500",
		"502",
		"503",
		"504",
		"timeout",
		"connection refused",
		"network error",
		"connection reset",
		"EOF",
	}

	for _, phrase := range retryablePhrases {
		if contains(errStr, phrase) {
			return true
		}
	}

	return false
}

func (g *HTTPOrderGatewayWithRetry) calculateDelay(attempt int) time.Duration {
	delay := float64(g.retryConfig.InitialDelay) * pow(g.retryConfig.BackoffFactor, float64(attempt))

	if delay > float64(g.retryConfig.MaxDelay) {
		delay = float64(g.retryConfig.MaxDelay)
	}

	if g.retryConfig.Jitter {
		delay = delay * (0.8 + 0.4*rand.Float64())
	}

	return time.Duration(delay)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func pow(x, y float64) float64 {
	if y == 0 {
		return 1
	}
	result := x
	for i := 1; i < int(y); i++ {
		result *= x
	}
	return result
}

func getErrorCode(err error) string {
	errStr := err.Error()
	if len(errStr) > 50 {
		return errStr[:50]
	}
	return errStr
}
