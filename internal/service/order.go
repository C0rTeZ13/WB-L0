package service

import (
	"L0/internal/cache"
	"L0/internal/kafka/dto"
	"L0/internal/repository"
	"context"
	"fmt"
	"log/slog"

	gocache "github.com/patrickmn/go-cache"
)

type orderService struct {
	repo   repository.Repository
	cache  cache.Cache
	logger *slog.Logger
}

func NewOrderService(repo repository.Repository, cache cache.Cache, logger *slog.Logger) OrderService {
	return &orderService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *orderService) GetOrder(ctx context.Context, orderUID string) (*dto.OrderDTO, error) {
	// Сначала пытаемся получить из кеша
	if cachedOrder, found := s.cache.Get(orderUID); found {
		if order, ok := cachedOrder.(*dto.OrderDTO); ok {
			s.logger.Debug("Order found in cache", slog.String("order_uid", orderUID))
			return order, nil
		}
		// Если в кеше лежит не тот тип, удаляем и идем в БД
		s.cache.Delete(orderUID)
		s.logger.Warn("Invalid type in cache, removed", slog.String("order_uid", orderUID))
	}

	// Если не в кеше, идем в БД
	s.logger.Info("Order not found in cache, fetching from database", slog.String("order_uid", orderUID))

	order, err := s.repo.GetOrderByUID(ctx, orderUID)
	if err != nil {
		s.logger.Error("Failed to get order from database", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Добавляем в кеш для будущих запросов
	s.cache.Set(orderUID, order, gocache.DefaultExpiration)
	s.logger.Debug("Order cached", slog.String("order_uid", orderUID))

	return order, nil
}

func (s *orderService) CreateOrder(ctx context.Context, order *dto.OrderDTO) (*dto.OrderDTO, error) {
	// Создаем заказ в БД
	createdOrder, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Error("Failed to create order", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Добавляем в кеш
	s.cache.Set(createdOrder.OrderUID, createdOrder, gocache.DefaultExpiration)
	s.logger.Info("Order created and cached", slog.String("order_uid", createdOrder.OrderUID))

	return createdOrder, nil
}
