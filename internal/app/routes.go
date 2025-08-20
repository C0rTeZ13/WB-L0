package app

import (
	"L0/internal/handlers"
	"L0/internal/service"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r *chi.Mux, orderService service.OrderService, logger *slog.Logger) {
	orderHandler := handlers.NewOrderHandler(orderService, logger)
	r.Route("/orders", func(r chi.Router) {
		r.Get("/{order_uid}", orderHandler.ServeHTTP)
	})
}
