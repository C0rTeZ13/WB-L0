package handlers

import (
	"L0/internal/service"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type OrderHandler struct {
	OrderService service.OrderService
	Logger       *slog.Logger
}

func NewOrderHandler(orderService service.OrderService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		OrderService: orderService,
		Logger:       logger,
	}
}

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orderUID := chi.URLParam(r, "order_uid")
	if orderUID == "" {
		h.Logger.Error("order_uid parameter is missing")
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	// Validate order_uid format
	if len(orderUID) == 0 || len(orderUID) > 100 {
		h.Logger.Error("Invalid order_uid format", slog.String("order_uid", orderUID))
		http.Error(w, "order_uid must be between 1 and 100 characters", http.StatusBadRequest)
		return
	}

	order, err := h.OrderService.GetOrder(r.Context(), orderUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.Logger.Info("Order not found", slog.String("order_uid", orderUID))
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}

		h.Logger.Error("Failed to get order", slog.String("error", err.Error()), slog.String("order_uid", orderUID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		h.Logger.Error("Failed to encode response", slog.String("error", err.Error()))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	h.Logger.Debug("Order retrieved successfully", slog.String("order_uid", orderUID))
}
