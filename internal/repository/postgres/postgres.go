package postgres

import (
	"L0/internal/kafka/dto"
	"L0/internal/models"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Storage struct {
	DB *gorm.DB
}

func (s *Storage) CreateOrder(ctx context.Context, o *dto.OrderDTO) (*dto.OrderDTO, error) {
	order, err := CreateOrder(ctx, s.DB, o)
	if err != nil {
		return nil, err
	}
	return convertToDTO(order), nil
}

func (s *Storage) GetAllOrders(ctx context.Context) ([]dto.OrderDTO, error) {
	return GetAllOrders(ctx, s.DB)
}

func (s *Storage) GetOrderByUID(ctx context.Context, orderUID string) (*dto.OrderDTO, error) {
	return GetOrderByUID(ctx, s.DB, orderUID)
}

// Close closes the database connection
func (s *Storage) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

func New(host string, port int, user, password, dbname string) (*Storage, error) {
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
		"file://internal/repository/postgres/migrations",
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
		"file://internal/repository/postgres/migrations",
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

func convertToDTO(order *models.Order) *dto.OrderDTO {
	if order == nil {
		return nil
	}

	o := &dto.OrderDTO{
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.ShardKey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated.Format(time.RFC3339),
		OofShard:          order.OofShard,
	}

	if order.Delivery.ID != 0 {
		o.Delivery = dto.DeliveryDTO{
			Name:    order.Delivery.Name,
			Phone:   order.Delivery.Phone,
			Zip:     order.Delivery.Zip,
			City:    order.Delivery.City,
			Address: order.Delivery.Address,
			Region:  order.Delivery.Region,
			Email:   order.Delivery.Email,
		}
	}

	if order.Payment.ID != 0 {
		o.Payment = dto.PaymentDTO{
			Transaction:  order.Payment.Transaction,
			RequestID:    order.Payment.RequestID,
			Currency:     order.Payment.Currency,
			Provider:     order.Payment.Provider,
			Amount:       order.Payment.Amount,
			PaymentDt:    order.Payment.PaymentDT,
			Bank:         order.Payment.Bank,
			DeliveryCost: order.Payment.DeliveryCost,
			GoodsTotal:   order.Payment.GoodsTotal,
			CustomFee:    order.Payment.CustomFee,
		}
	}

	for _, item := range order.Items {
		o.Items = append(o.Items, dto.ItemDTO{
			ChrtID:      item.ChrtID,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			RID:         item.RID,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmID:        item.NmID,
			Brand:       item.Brand,
			Status:      item.Status,
		})
	}

	return o
}
