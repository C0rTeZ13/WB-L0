package kafka

import (
	"L0/internal/repository"
	"context"
)

// Consumer интерфейс для потребления сообщений из Kafka
type Consumer interface {
	ConsumeOrders(ctx context.Context, brokers []string, topic, groupID string, repo repository.Repository) error
}

// MessageProcessor интерфейс для обработки сообщений
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, data []byte, repo repository.Repository) error
}
