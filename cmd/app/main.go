package main

import (
	"L0/internal/api"
	"L0/internal/config"
	"L0/internal/kafka"
	"L0/internal/storage/postgres"
	"context"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var cache sync.Map

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Starting app", slog.String("env", cfg.Env))

	storage, err := postgres.New(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Error("Failed init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := loadCacheFromDB(storage); err != nil {
		log.Error("Failed to load cache from DB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go kafka.ConsumeOrders(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, storage)

	r := chi.NewRouter()
	r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.yaml"),
	))

	api.RegisterRoutes(r, storage, log, &cache)

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Error("Failed serve", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func loadCacheFromDB(storage *postgres.Storage) error {
	orders, err := postgres.GetAllOrders(context.Background(), storage.Client)
	if err != nil {
		return err
	}
	for _, order := range orders {
		cache.Store(order.OrderUID, order)
	}
	return nil
}
