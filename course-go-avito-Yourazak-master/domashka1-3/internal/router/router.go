package router

import (
	"log"
	"net/http"

	"avito-courier/internal/handler"
	"avito-courier/internal/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(courierHandler *handler.CourierHandler, deliveryHandler *handler.DeliveryHandler, rateLimiter *middleware.RateLimiter) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/couriers", courierHandler.Create)
	mux.HandleFunc("GET /api/couriers/{id}", courierHandler.GetByID)
	mux.HandleFunc("PUT /api/couriers/{id}", courierHandler.Update)
	mux.HandleFunc("DELETE /api/couriers/{id}", courierHandler.Delete)
	mux.HandleFunc("GET /api/couriers", courierHandler.GetAll)
	mux.HandleFunc("POST /api/courier/assign", courierHandler.AssignOrder)

	mux.HandleFunc("POST /api/delivery/assign", deliveryHandler.Assign)
	mux.HandleFunc("POST /api/delivery/unassign", deliveryHandler.Unassign)
	mux.HandleFunc("GET /api/delivery/{id}", deliveryHandler.GetDelivery)
	mux.HandleFunc("GET /api/deliveries", deliveryHandler.ListDeliveries)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle("/metrics", promhttp.Handler())

	handler := http.Handler(mux)

	handler = middleware.MetricsMiddleware(handler)

	handler = middleware.LoggingMiddleware(handler)

	if rateLimiter != nil {
		handler = rateLimiter.Middleware()(handler)
		log.Println(" Rate limiter middleware applied")
	}

	return handler
}
