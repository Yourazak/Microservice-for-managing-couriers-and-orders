package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"avito-courier/internal/model"
	"avito-courier/internal/usecase"
)

type CourierHandler struct {
	courierUC usecase.CourierUsecase
}

func NewCourierHandler(courierUC usecase.CourierUsecase) *CourierHandler {
	return &CourierHandler{courierUC: courierUC}
}

func (h *CourierHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var courier model.Courier
	if err := json.NewDecoder(r.Body).Decode(&courier); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.courierUC.Create(r.Context(), &courier); err != nil {
		switch err {
		case usecase.ErrBadInput:
			http.Error(w, "Invalid input data", http.StatusBadRequest)
		case usecase.ErrConflict:
			http.Error(w, "Courier already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(courier)
}

func (h *CourierHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/couriers/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	courier, err := h.courierUC.GetByID(r.Context(), id)
	if err != nil {
		if err == usecase.ErrNotFound {
			http.Error(w, "Courier not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courier)
}

func (h *CourierHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/couriers/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var courier model.Courier
	if err := json.NewDecoder(r.Body).Decode(&courier); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	courier.ID = id
	if err := h.courierUC.Update(r.Context(), &courier); err != nil {
		switch err {
		case usecase.ErrNotFound:
			http.Error(w, "Courier not found", http.StatusNotFound)
		case usecase.ErrBadInput:
			http.Error(w, "Invalid input data", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courier)
}

func (h *CourierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/couriers/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	http.Error(w, "Delete not implemented", http.StatusNotImplemented)
}

func (h *CourierHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	couriers, err := h.courierUC.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(couriers)
}

func (h *CourierHandler) AssignOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID string `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OrderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	http.Error(w, "Assign not implemented in courier handler", http.StatusNotImplemented)
}
