package handlers

import (
	"L0/ent"
	"L0/internal/storage/postgres"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
)

type OrderHandler struct {
	Storage *postgres.Storage
	Logger  *slog.Logger
	Cache   *sync.Map
}

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "order_id")
	if orderIDStr == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	}

	cachedOrder, found := h.Cache.Load(orderID)
	if found {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cachedOrder); err != nil {
			h.Logger.Error("failed to encode cached order", slog.String("error", err.Error()))
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	ctx := r.Context()
	order, err := h.Storage.Client.Order.Get(ctx, orderID)
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		h.Logger.Error("failed to get order", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.Cache.Store(orderID, order)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		h.Logger.Error("failed to encode response", slog.String("error", err.Error()))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
