package kafka

import (
	"L0/internal/kafka/dto"
	"L0/internal/storage/postgres"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

func ConsumeOrders(ctx context.Context, brokers []string, topic, groupID string, storage *postgres.Storage) error {
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
			log.Printf("connection error: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer stopped")
			return nil
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				log.Printf("error reading message: %v", err)
				continue
			}

			var order dto.OrderDTO
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("invalid message format: %v", err)
				_ = r.CommitMessages(ctx, m)
				continue
			}

			for {
				_, err := postgres.CreateOrder(ctx, storage.Client, &order)
				if err != nil {
					var pgErr *pq.Error
					if errors.As(err, &pgErr) && pgErr.Code == "23505" {
						log.Printf("order UID %s already exists, skipping retry. Table: %s, Constraint: %s",
							order.OrderUID, pgErr.Table, pgErr.Constraint)
						break
					}

					log.Printf("failed to save order UID %s: %v, retrying...", order.OrderUID, err)
					select {
					case <-ctx.Done():
						return nil
					case <-time.After(3 * time.Second):
						continue
					}
				}
				break
			}

			if err := r.CommitMessages(ctx, m); err != nil {
				log.Printf("failed to commit message: %v", err)
			}
		}
	}
}
