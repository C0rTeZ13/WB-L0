package postgres

import (
	"L0/ent"
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"log/slog"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Storage struct {
	Client *ent.Client
}

func New(host string, port int, user, password, dbname string) (*Storage, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	drv, err := sql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening postgres driver: %w", err)
	}

	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	return &Storage{Client: client}, nil
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
