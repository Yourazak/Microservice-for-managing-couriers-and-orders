package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"avito-courier/internal/config"
	"avito-courier/internal/model"
)

type OrderGateway interface {
	GetOrdersByCursor(ctx context.Context, cursor time.Time) ([]model.OrderEvent, error)
	GetOrderStatus(ctx context.Context, orderID string) (*model.OrderEvent, error)
}

type HTTPOrderGateway struct {
	client  *http.Client
	baseURL string
}

func NewHTTPOrderGateway(cfg *config.Config) *HTTPOrderGateway {
	return &HTTPOrderGateway{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: cfg.ServiceOrderURL,
	}
}

func (g *HTTPOrderGateway) GetOrdersByCursor(ctx context.Context, cursor time.Time) ([]model.OrderEvent, error) {
	u, err := url.Parse(g.baseURL + "/public/api/v1/orders")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("from", cursor.Format(time.RFC3339))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var orders []model.OrderEvent
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return orders, nil
}

func (g *HTTPOrderGateway) GetOrderStatus(ctx context.Context, orderID string) (*model.OrderEvent, error) {
	url := fmt.Sprintf("%s/public/api/v1/order/%s", g.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var order model.OrderEvent
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &order, nil
}
