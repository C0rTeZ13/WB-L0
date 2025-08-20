package service_test

import (
	"L0/internal/kafka/dto"
	"L0/internal/service"
	"L0/test/mocks"
	"L0/test/testutils"
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderService_GetOrder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name          string
		orderUID      string
		setupMocks    func(*mocks.MockRepository, *mocks.MockCache)
		expectedError string
		expectedCalls map[string]int
	}{
		{
			name:     "success_from_cache",
			orderUID: "test_order_1",
			setupMocks: func(repo *mocks.MockRepository, cache *mocks.MockCache) {
				order := testutils.MinimalOrderFixture("test_order_1")
				cache.Store["test_order_1"] = order
			},
			expectedError: "",
			expectedCalls: map[string]int{
				"cache_get":    1,
				"repo_get":     0,
				"cache_set":    0,
				"cache_delete": 0,
			},
		},
		{
			name:     "success_from_repo",
			orderUID: "test_order_2",
			setupMocks: func(repo *mocks.MockRepository, cache *mocks.MockCache) {
				order := testutils.MinimalOrderFixture("test_order_2")
				_, err := repo.CreateOrder(context.Background(), order)
				if err != nil {
					require.NoError(t, err, "Failed to create test order in mock repository")
				}
			},
			expectedError: "",
			expectedCalls: map[string]int{
				"cache_get":    1,
				"repo_get":     1,
				"cache_set":    1,
				"cache_delete": 0,
			},
		},
		{
			name:     "order_not_found",
			orderUID: "nonexistent_order",
			setupMocks: func(repo *mocks.MockRepository, cache *mocks.MockCache) {
				// Nothing
			},
			expectedError: "failed to get order",
			expectedCalls: map[string]int{
				"cache_get": 1,
				"repo_get":  1,
				"cache_set": 0,
			},
		},
		{
			name:     "repo_error",
			orderUID: "test_order_3",
			setupMocks: func(repo *mocks.MockRepository, cache *mocks.MockCache) {
				repo.ShouldFail = true
				repo.FailError = errors.New("database error")
			},
			expectedError: "failed to get order",
			expectedCalls: map[string]int{
				"cache_get": 1,
				"repo_get":  1,
				"cache_set": 0,
			},
		},
		{
			name:     "invalid_cache_data",
			orderUID: "test_order_4",
			setupMocks: func(repo *mocks.MockRepository, cache *mocks.MockCache) {
				cache.Store["test_order_4"] = "invalid_data"
				order := testutils.MinimalOrderFixture("test_order_4")
				_, err := repo.CreateOrder(context.Background(), order)
				if err != nil {
					require.NoError(t, err, "Failed to create test order in mock repository")
				}
			},
			expectedError: "",
			expectedCalls: map[string]int{
				"cache_get":    1,
				"cache_delete": 1,
				"repo_get":     1,
				"cache_set":    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			mockCache := mocks.NewMockCache()
			tt.setupMocks(mockRepo, mockCache)

			orderService := service.NewOrderService(mockRepo, mockCache, logger)

			result, err := orderService.GetOrder(context.Background(), tt.orderUID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.orderUID, result.OrderUID)
			}

			assert.Equal(t, tt.expectedCalls["cache_get"], mockCache.CallsGet, "cache Get calls")
			assert.Equal(t, tt.expectedCalls["repo_get"], mockRepo.CallsGetOrderByUID, "repo GetOrderByUID calls")
			assert.Equal(t, tt.expectedCalls["cache_set"], mockCache.CallsSet, "cache Set calls")

			if deleteCount, ok := tt.expectedCalls["cache_delete"]; ok {
				assert.Equal(t, deleteCount, mockCache.CallsDelete, "cache Delete calls")
			}
		})
	}
}

func TestOrderService_CreateOrder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name          string
		order         func() *dto.OrderDTO
		setupMocks    func(*mocks.MockRepository, *mocks.MockCache)
		expectedError string
	}{
		{
			name: "success",
			order: func() *dto.OrderDTO {
				return testutils.MinimalOrderFixture("success_order")
			},
			setupMocks: func(repo *mocks.MockRepository, cache *mocks.MockCache) {
				// Nothing
			},
			expectedError: "",
		},
		{
			name: "repo_error",
			order: func() *dto.OrderDTO {
				return testutils.MinimalOrderFixture("error_order")
			},
			setupMocks: func(repo *mocks.MockRepository, cache *mocks.MockCache) {
				repo.ShouldFail = true
				repo.FailError = errors.New("database connection failed")
			},
			expectedError: "failed to create order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			mockCache := mocks.NewMockCache()
			tt.setupMocks(mockRepo, mockCache)

			orderService := service.NewOrderService(mockRepo, mockCache, logger)
			testOrder := tt.order()

			result, err := orderService.CreateOrder(context.Background(), testOrder)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)

				assert.Equal(t, 0, mockCache.CallsSet)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, testOrder.OrderUID, result.OrderUID)

				assert.Equal(t, 1, mockCache.CallsSet)
				assert.Equal(t, 1, mockRepo.CallsCreateOrder)
			}
		})
	}
}

func TestOrderService_GetOrder_ConcurrentAccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	mockRepo := mocks.NewMockRepository()
	mockCache := mocks.NewMockCache()

	order := testutils.MinimalOrderFixture("concurrent_test")
	_, err := mockRepo.CreateOrder(context.Background(), order)
	if err != nil {
		require.NoError(t, err, "Failed to create test order in mock repository")
	}

	orderService := service.NewOrderService(mockRepo, mockCache, logger)

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result, err := orderService.GetOrder(context.Background(), "concurrent_test")
			if err != nil {
				results <- err
				return
			}
			if result == nil || result.OrderUID != "concurrent_test" {
				results <- errors.New("invalid result")
				return
			}
			results <- nil
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			assert.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	assert.Greater(t, mockCache.CallsSet, 0, "Cache should be set at least once")
}
