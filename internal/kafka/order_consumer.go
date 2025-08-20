package kafka

import (
	"L0/internal/kafka/dto"
	"L0/internal/repository"
	"L0/internal/validation"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type orderConsumer struct {
	logger *slog.Logger
}

type orderMessageProcessor struct {
	logger *slog.Logger
}

// NewOrderConsumer создает новый экземпляр Kafka consumer для заказов
func NewOrderConsumer(logger *slog.Logger) Consumer {
	return &orderConsumer{
		logger: logger,
	}
}

// NewOrderMessageProcessor создает новый экземпляр процессора сообщений для заказов
func NewOrderMessageProcessor(logger *slog.Logger) MessageProcessor {
	return &orderMessageProcessor{
		logger: logger,
	}
}

func (c *orderConsumer) ConsumeOrders(ctx context.Context, brokers []string, topic, groupID string, repo repository.Repository) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 0,
	})

	defer func() {
		if err := r.Close(); err != nil {
			c.logger.Error("Failed to close Kafka connection", slog.String("error", err.Error()))
		}
	}()

	c.logger.Info("Kafka consumer started",
		slog.String("topic", topic),
		slog.String("groupID", groupID),
		slog.Any("brokers", brokers))

	processor := NewOrderMessageProcessor(c.logger)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Kafka consumer stopped")
			return nil
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				c.logger.Error("Error reading Kafka message", slog.String("error", err.Error()))
				continue
			}

			if err := processor.ProcessMessage(ctx, m.Value, repo); err != nil {
				c.logger.Error("Failed to process message", slog.String("error", err.Error()))
				// Пропускаем сообщение, если не удалось его обработать
			}

			if err := r.CommitMessages(ctx, m); err != nil {
				c.logger.Error("Failed to commit Kafka message", slog.String("error", err.Error()))
			}
		}
	}
}

func (p *orderMessageProcessor) ProcessMessage(ctx context.Context, data []byte, repo repository.Repository) error {
	var order dto.OrderDTO
	if err := json.Unmarshal(data, &order); err != nil {
		return errors.New("invalid message format: " + err.Error())
	}

	// Validate order data
	validator := validation.NewOrderValidator()
	if err := validator.ValidateOrder(&order); err != nil {
		p.logger.Error("Order validation failed",
			slog.String("order_uid", order.OrderUID),
			slog.String("validation_error", err.Error()))
		return errors.New("order validation failed: " + err.Error())
	}

	// Retry loop for database operations
	for {
		_, err := repo.CreateOrder(ctx, &order)
		if err != nil {
			var pgErr *pq.Error
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				p.logger.Info("Order already exists, skipping",
					slog.String("order_uid", order.OrderUID),
					slog.String("table", pgErr.Table),
					slog.String("constraint", pgErr.Constraint))
				return nil
			}

			p.logger.Error("Failed to save order, retrying...",
				slog.String("order_uid", order.OrderUID),
				slog.String("error", err.Error()))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(3 * time.Second):
				continue
			}
		}
		break
	}

	p.logger.Info("Order processed successfully", slog.String("order_uid", order.OrderUID))
	return nil
}
