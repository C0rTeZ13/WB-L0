package main

import (
	"L0/internal/config"
	"L0/internal/repository/postgres"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	if len(os.Args) < 2 {
		log.Error("No command specified", slog.String("usage", "migrate up|down"))
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "up":
		if err := postgres.RunMigrations(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName); err != nil {
			log.Error("Failed to run migrations", slog.String("error", err.Error()))
			os.Exit(1)
		}
		log.Info("Migrations applied successfully")
	case "down":
		if err := postgres.RollbackLast(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName); err != nil {
			log.Error("Failed to rollback migration", slog.String("error", err.Error()))
			os.Exit(1)
		}
		log.Info("Rollback applied successfully")
	default:
		log.Error("Unknown command", slog.String("usage", "migrate up|down"))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return log
}
