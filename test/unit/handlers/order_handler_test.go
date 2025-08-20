package handlers_test

import (
	"L0/internal/handlers"
	"L0/test/mocks"
	"L0/test/testutils"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOrderHandler_ServeHTTP(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name               string
		orderUID           string
		setupMockService   func(*mocks.MockOrderService)
		expectedStatusCode int
		expectedResponse   string
		expectJSON         bool
	}{
		{
			name:     "success_order_found",
			orderUID: "test_order_123",
			setupMockService: func(mockService *mocks.MockOrderService) {
				order := testutils.MinimalOrderFixture("test_order_123")
				mockService.AddOrder(order)
			},
			expectedStatusCode: http.StatusOK,
			expectJSON:         true,
		},
		{
			name:     "order_not_found",
			orderUID: "nonexistent_order",
			setupMockService: func(mockService *mocks.MockOrderService) {
				mockService.ShouldFail = true
				mockService.FailError = gorm.ErrRecordNotFound
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   "order not found",
			expectJSON:         false,
		},
		{
			name:     "internal_server_error",
			orderUID: "error_order",
			setupMockService: func(mockService *mocks.MockOrderService) {
				mockService.ShouldFail = true
				mockService.FailError = errors.New("database connection failed")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "internal server error",
			expectJSON:         false,
		},
		{
			name:     "missing_order_uid",
			orderUID: "", // Пустой UID
			setupMockService: func(mockService *mocks.MockOrderService) {
				// Не важно, сервис не будет вызван
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "order_uid is required",
			expectJSON:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := mocks.NewMockOrderService()
			tt.setupMockService(mockService)

			handler := handlers.NewOrderHandler(mockService, logger)

			// Создаем HTTP запрос
			req := httptest.NewRequest("GET", "/orders/"+tt.orderUID, nil)
			if tt.orderUID != "" {
				// Добавляем URL параметр для chi router
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("order_uid", tt.orderUID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			recorder := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(recorder, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, recorder.Code)

			if tt.expectJSON {
				assert.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
				assert.Contains(t, recorder.Body.String(), tt.orderUID)
			} else if tt.expectedResponse != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedResponse)
			}

			// Проверяем, что сервис был вызван правильное количество раз
			if tt.orderUID != "" && recorder.Code != http.StatusBadRequest {
				assert.Equal(t, 1, mockService.CallsGetOrder)
			} else {
				assert.Equal(t, 0, mockService.CallsGetOrder)
			}
		})
	}
}

func TestOrderHandler_ServeHTTP_Integration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Создаем мок сервис с тестовыми данными
	mockService := mocks.NewMockOrderService()
	order1 := testutils.MinimalOrderFixture("order_1")
	order2 := testutils.OrderFixture() // Полный заказ
	mockService.AddOrder(order1)
	mockService.AddOrder(order2)

	handler := handlers.NewOrderHandler(mockService, logger)

	// Создаем роутер как в реальном приложении
	r := chi.NewRouter()
	r.Get("/orders/{order_uid}", handler.ServeHTTP)

	tests := []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedOrderUID   string
	}{
		{
			name:               "get_order_1",
			url:                "/orders/order_1",
			expectedStatusCode: http.StatusOK,
			expectedOrderUID:   "order_1",
		},
		{
			name:               "get_full_order",
			url:                "/orders/" + order2.OrderUID,
			expectedStatusCode: http.StatusOK,
			expectedOrderUID:   order2.OrderUID,
		},
		{
			name:               "order_not_found",
			url:                "/orders/unknown_order",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			recorder := httptest.NewRecorder()

			r.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Code)

			if tt.expectedOrderUID != "" {
				assert.Contains(t, recorder.Body.String(), tt.expectedOrderUID)
				assert.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
			}
		})
	}
}

func TestOrderHandler_ServeHTTP_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("service_timeout", func(t *testing.T) {
		mockService := mocks.NewMockOrderService()
		mockService.ShouldFail = true
		mockService.FailError = context.DeadlineExceeded

		handler := handlers.NewOrderHandler(mockService, logger)

		req := httptest.NewRequest("GET", "/orders/timeout_order", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("order_uid", "timeout_order")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "internal server error")
	})

	t.Run("service_panic_recovery", func(t *testing.T) {
		// Этот тест проверяет, что хендлер не паникует
		// даже если сервис возвращает nil результат
		mockService := mocks.NewMockOrderService()

		handler := handlers.NewOrderHandler(mockService, logger)

		req := httptest.NewRequest("GET", "/orders/panic_order", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("order_uid", "panic_order")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		recorder := httptest.NewRecorder()

		// Должно выполниться без паники
		require.NotPanics(t, func() {
			handler.ServeHTTP(recorder, req)
		})

		// Ожидаем ошибку, так как заказ не найден
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
}
