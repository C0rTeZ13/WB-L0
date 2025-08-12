package handlers

import (
	"L0/ent"
	"L0/ent/order"
	"L0/internal/storage/postgres"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"sync"
)

type OrderHandler struct {
	Storage *postgres.Storage
	Logger  *slog.Logger
	Cache   *sync.Map
}

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orderUID := chi.URLParam(r, "order_uid")
	if orderUID == "" {
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	cachedOrder, found := h.Cache.Load(orderUID)
	if found {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cachedOrder); err != nil {
			h.Logger.Error("failed to encode cached order", slog.String("error", err.Error()))
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	ctx := r.Context()
	foundOrder, err := h.Storage.Client.Order.
		Query().
		Where(order.OrderUIDEQ(orderUID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		h.Logger.Error("failed to get order", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.Cache.Store(orderUID, foundOrder)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(foundOrder); err != nil {
		h.Logger.Error("failed to encode response", slog.String("error", err.Error()))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
