package kafka

import (
	"L0/internal/kafka/dto"
	"L0/internal/storage/postgres"
	"context"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func ConsumeOrders(brokers []string, topic, groupID string, storage *postgres.Storage) {
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
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		var order dto.OrderDTO
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("invalid message format: %v", err)
			_ = r.CommitMessages(context.Background(), m)
			continue
		}

		for {
			_, err := postgres.CreateOrder(context.Background(), storage.Client, &order)
			if err != nil {
				var pgErr *pq.Error
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					log.Printf("order UID %s already exists, skipping retry. Table: %s, Constraint: %s", order.OrderUID, pgErr.Table, pgErr.Constraint)
					break
				}

				log.Printf("failed to save order UID %s: %v, retrying...", order.OrderUID, err)
				time.Sleep(time.Second * 3)
				continue
			}
			break
		}

		if err := r.CommitMessages(context.Background(), m); err != nil {
			log.Printf("failed to commit message: %v", err)
		}
	}
}
