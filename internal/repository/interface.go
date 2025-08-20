package repository

import (
	"L0/internal/kafka/dto"
	"context"
)

type Repository interface {
	CreateOrder(ctx context.Context, o *dto.OrderDTO) (*dto.OrderDTO, error)
	GetAllOrders(ctx context.Context) ([]dto.OrderDTO, error)
	GetOrderByUID(ctx context.Context, orderUID string) (*dto.OrderDTO, error)
}
