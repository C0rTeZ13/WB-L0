package service

import (
	"L0/internal/kafka/dto"
	"context"
)

type OrderService interface {
	GetOrder(ctx context.Context, orderUID string) (*dto.OrderDTO, error)
	CreateOrder(ctx context.Context, order *dto.OrderDTO) (*dto.OrderDTO, error)
}
