package kafka

import (
	"L0/internal/kafka/dto"
	"L0/internal/storage/postgres"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
)

func ConsumeOrders(brokers []string, topic, groupID string, storage *postgres.Storage) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
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
			log.Printf("failed to unmarshal order: %v", err)
			if err := r.CommitMessages(context.Background(), m); err != nil {
				log.Printf("failed to commit message: %v", err)
			}
			continue
		}

		if _, err := postgres.CreateOrder(context.Background(), storage.Client, &order); err != nil {
			log.Printf("failed to save order UID %s: %v", order.OrderUID, err)
			continue
		}

		if err := r.CommitMessages(context.Background(), m); err != nil {
			log.Printf("failed to commit message: %v", err)
		}
	}
}
