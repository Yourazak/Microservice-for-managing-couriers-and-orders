package handler

import (
	"encoding/json"
	"net/http"

	"avito-courier/internal/model"
	"avito-courier/internal/usecase"
)

type DeliveryHandler struct {
	deliveryUC usecase.IDeliveryUsecase
}

func NewDeliveryHandler(deliveryUC usecase.IDeliveryUsecase) *DeliveryHandler {
	return &DeliveryHandler{deliveryUC: deliveryUC}
}

func (h *DeliveryHandler) Assign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID string `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OrderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	delivery, courier, err := h.deliveryUC.Assign(r.Context(), req.OrderID)
	if err != nil {
		switch err {
		case usecase.ErrNoAvailableCourier:
			http.Error(w, "No available couriers", http.StatusConflict)
		case usecase.ErrOrderAlreadyAssigned:
			http.Error(w, "Order already assigned", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := struct {
		Delivery model.Delivery `json:"delivery"`
		Courier  model.Courier  `json:"courier"`
	}{
		Delivery: delivery,
		Courier:  courier,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *DeliveryHandler) Unassign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID string `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OrderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	if err := h.deliveryUC.Unassign(r.Context(), req.OrderID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "unassigned"})
}

func (h *DeliveryHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"message": "Not implemented"})
}

func (h *DeliveryHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"message": "Not implemented"})
}
