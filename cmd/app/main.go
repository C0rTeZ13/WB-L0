package main

import (
	"L0/internal/api"
	"L0/internal/config"
	"L0/internal/kafka"
	"L0/internal/storage/postgres"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

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
	log.Info("Starting service", slog.String("env", cfg.Env))

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 2)

	go func() {
		errCh <- kafka.ConsumeOrders(ctx,
			cfg.Kafka.Brokers,
			cfg.Kafka.Topic,
			cfg.Kafka.GroupID,
			storage,
		)
	}()

	r := chi.NewRouter()
	r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.yaml"),
	))

	api.RegisterRoutes(r, storage, log, &cache)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Info("HTTP server started on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("Shutting down gracefully...")
	case err := <-errCh:
		if err != nil {
			log.Error("Service error", slog.String("error", err.Error()))
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Failed to shutdown HTTP server", slog.String("error", err.Error()))
	}

	log.Info("Service stopped")
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
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
