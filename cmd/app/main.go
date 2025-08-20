package main

import (
	"L0/internal/app"
	"L0/internal/cache"
	"L0/internal/config"
	"L0/internal/kafka"
	"L0/internal/repository"
	"L0/internal/repository/postgres"
	"L0/internal/service"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	gocache "github.com/patrickmn/go-cache"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("Starting service", slog.String("env", cfg.Env))

	cacheImpl := cache.New(5*time.Minute, 10*time.Minute)

	storageImpl, err := postgres.New(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Error("Failed init repository", slog.String("error", err.Error()))
		os.Exit(1)
	}
	var repo repository.Repository = storageImpl

	if err := loadCacheFromDB(storageImpl, cacheImpl); err != nil {
		log.Error("Failed to load cache from DB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	orderService := service.NewOrderService(repo, cacheImpl, log)

	kafkaConsumer := kafka.NewOrderConsumer(log)

	r := chi.NewRouter()
	r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/docs/swagger.yaml")))
	app.RegisterRoutes(r, orderService, log)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		log.Info("HTTP server started on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		if err := kafkaConsumer.ConsumeOrders(ctx, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, repo); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("Shutting down gracefully...")
	case err := <-errCh:
		log.Error("Service error", slog.String("error", err.Error()))
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Failed to shutdown HTTP server", slog.String("error", err.Error()))
	} else {
		log.Info("HTTP server shutdown successfully")
	}

	if err := storageImpl.Close(); err != nil {
		log.Error("Failed to close database connection", slog.String("error", err.Error()))
	} else {
		log.Info("Database connection closed successfully")
	}

	log.Info("Service stopped")
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case "local":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "dev":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

func loadCacheFromDB(storage *postgres.Storage, cache cache.Cache) error {
	orders, err := postgres.GetAllOrders(context.Background(), storage.DB)
	if err != nil {
		return err
	}
	for _, order := range orders {
		cache.Set(order.OrderUID, &order, gocache.DefaultExpiration)
	}
	return nil
}
