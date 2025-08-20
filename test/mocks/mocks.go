package mocks

import (
	"L0/internal/cache"
	"L0/internal/kafka/dto"
	"L0/internal/repository"
	"L0/internal/service"
	"context"
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"
)

// MockRepository - мок для repository.Repository
type MockRepository struct {
	mu     sync.RWMutex
	orders map[string]*dto.OrderDTO

	// Для контроля поведения
	ShouldFail         bool
	FailError          error
	CallsCreateOrder   int
	CallsGetOrderByUID int
	CallsGetAllOrders  int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		orders: make(map[string]*dto.OrderDTO),
	}
}

func (m *MockRepository) CreateOrder(ctx context.Context, order *dto.OrderDTO) (*dto.OrderDTO, error) {
	_ = ctx // Параметр не используется в моке, но необходим для соответствия интерфейсу
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallsCreateOrder++

	if m.ShouldFail {
		return nil, m.FailError
	}

	m.orders[order.OrderUID] = order
	return order, nil
}

func (m *MockRepository) GetOrderByUID(ctx context.Context, orderUID string) (*dto.OrderDTO, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.CallsGetOrderByUID++

	if m.ShouldFail {
		return nil, m.FailError
	}

	order, exists := m.orders[orderUID]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

func (m *MockRepository) GetAllOrders(ctx context.Context) ([]dto.OrderDTO, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.CallsGetAllOrders++

	if m.ShouldFail {
		return nil, m.FailError
	}

	result := make([]dto.OrderDTO, 0, len(m.orders))
	for _, order := range m.orders {
		result = append(result, *order)
	}

	return result, nil
}

// Reset сбрасывает состояние мока
func (m *MockRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders = make(map[string]*dto.OrderDTO)
	m.ShouldFail = false
	m.FailError = nil
	m.CallsCreateOrder = 0
	m.CallsGetOrderByUID = 0
	m.CallsGetAllOrders = 0
}

// MockCache - мок для cache.Cache
type MockCache struct {
	mu   sync.RWMutex
	data map[string]interface{}

	// Для контроля поведения
	ShouldFail  bool
	CallsGet    int
	CallsSet    int
	CallsDelete int

	// Store - публичное поле для прямого доступа к данным в тестах
	Store map[string]interface{}
}

func NewMockCache() *MockCache {
	c := &MockCache{
		data:  make(map[string]interface{}),
		Store: make(map[string]interface{}),
	}
	c.Store = c.data
	return c
}

func (m *MockCache) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.CallsGet++

	if m.ShouldFail {
		return nil, false
	}

	value, exists := m.data[key]
	return value, exists
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) {
	_ = ttl
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallsSet++

	if !m.ShouldFail {
		m.data[key] = value
	}
}

func (m *MockCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallsDelete++

	if !m.ShouldFail {
		delete(m.data, key)
	}
}

// Reset сбрасывает состояние мока
func (m *MockCache) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]interface{})
	m.Store = m.data // Синхронизируем Store с data
	m.ShouldFail = false
	m.CallsGet = 0
	m.CallsSet = 0
	m.CallsDelete = 0
}

// MockOrderService - мок для service.OrderService
type MockOrderService struct {
	mu     sync.RWMutex
	orders map[string]*dto.OrderDTO

	// Для контроля поведения
	ShouldFail       bool
	FailError        error
	CallsGetOrder    int
	CallsCreateOrder int
}

func NewMockOrderService() *MockOrderService {
	return &MockOrderService{
		orders: make(map[string]*dto.OrderDTO),
	}
}

func (m *MockOrderService) GetOrder(ctx context.Context, orderUID string) (*dto.OrderDTO, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.CallsGetOrder++

	if m.ShouldFail {
		return nil, m.FailError
	}

	order, exists := m.orders[orderUID]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}

	return order, nil
}

func (m *MockOrderService) CreateOrder(ctx context.Context, order *dto.OrderDTO) (*dto.OrderDTO, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallsCreateOrder++

	if m.ShouldFail {
		return nil, m.FailError
	}

	m.orders[order.OrderUID] = order
	return order, nil
}

// AddOrder добавляет заказ в мок для тестирования
func (m *MockOrderService) AddOrder(order *dto.OrderDTO) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[order.OrderUID] = order
}

// Reset сбрасывает состояние мока
func (m *MockOrderService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders = make(map[string]*dto.OrderDTO)
	m.ShouldFail = false
	m.FailError = nil
	m.CallsGetOrder = 0
	m.CallsCreateOrder = 0
}

// Проверяем, что моки реализуют интерфейсы
var (
	_ repository.Repository = (*MockRepository)(nil)
	_ cache.Cache           = (*MockCache)(nil)
	_ service.OrderService  = (*MockOrderService)(nil)
)
