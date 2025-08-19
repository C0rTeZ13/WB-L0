package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Storage struct {
	DB *gorm.DB
}

func New(ctx context.Context, host string, port int, user, password, dbname string) (*Storage, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed opening postgres connection: %w", err)
	}

	return &Storage{DB: db}, nil
}

func RunMigrations(host string, port int, user, password, dbname string) error {
	m, err := migrate.New(
		"file://internal/storage/postgres/migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			user, password, host, port, dbname),
	)
	if err != nil {
		return err
	}
	defer func() {
		if cerr, _ := m.Close(); cerr != nil {
			slog.Error("failed to close migrate", slog.String("error", cerr.Error()))
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func RollbackLast(host string, port int, user, password, dbname string) error {
	m, err := migrate.New(
		"file://internal/storage/postgres/migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			user, password, host, port, dbname),
	)
	if err != nil {
		return err
	}
	defer func() {
		if cerr, _ := m.Close(); cerr != nil {
			slog.Error("failed to close migrate", slog.String("error", cerr.Error()))
		}
	}()

	return m.Steps(-1)
}
