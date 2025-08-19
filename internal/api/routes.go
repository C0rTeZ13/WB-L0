package api

import (
	"L0/internal/handlers"
	"L0/internal/storage/postgres"
	"log/slog"
	"sync"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r *chi.Mux, storage *postgres.Storage, logger *slog.Logger, cache *sync.Map) {
	orderHandler := &handlers.OrderHandler{Storage: storage, Logger: logger, Cache: cache}
	r.Route("/orders", func(r chi.Router) {
		r.Get("/{order_uid}", orderHandler.ServeHTTP)
	})
}
