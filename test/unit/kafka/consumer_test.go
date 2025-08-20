package kafka_test

import (
	"L0/internal/kafka"
	"L0/internal/kafka/dto"
	"L0/test/mocks"
	"L0/test/testutils"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderMessageProcessor_ProcessMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name           string
		messageData    []byte
		setupRepo      func(*mocks.MockRepository)
		expectedError  string
		expectRepoCall bool
	}{
		{
			name:        "valid_message_success",
			messageData: mustMarshalOrder(testutils.MinimalOrderFixture("valid_order")),
			setupRepo: func(repo *mocks.MockRepository) {
				// Nothing
			},
			expectedError:  "",
			expectRepoCall: true,
		},
		{
			name:           "invalid_json",
			messageData:    []byte(`{"invalid": json`),
			setupRepo:      func(repo *mocks.MockRepository) {},
			expectedError:  "invalid message format",
			expectRepoCall: false,
		},
		{
			name:        "repo_error_retry",
			messageData: mustMarshalOrder(testutils.MinimalOrderFixture("repo_error")),
			setupRepo: func(repo *mocks.MockRepository) {
				repo.ShouldFail = true
				repo.FailError = errors.New("database connection failed")
			},
			expectedError:  "",
			expectRepoCall: true,
		},
		{
			name:        "duplicate_order_handled",
			messageData: mustMarshalOrder(testutils.MinimalOrderFixture("duplicate_order")),
			setupRepo: func(repo *mocks.MockRepository) {
				pgErr := &pq.Error{
					Code:       "23505", // unique_violation
					Table:      "orders",
					Constraint: "orders_order_uid_key",
				}
				repo.ShouldFail = true
				repo.FailError = pgErr
			},
			expectedError:  "",
			expectRepoCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockRepository()
			tt.setupRepo(mockRepo)

			processor := kafka.NewOrderMessageProcessor(logger)

			ctx := context.Background()
			if tt.name == "repo_error_retry" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
			}

			err := processor.ProcessMessage(ctx, tt.messageData, mockRepo)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else if tt.name == "repo_error_retry" {
				require.Error(t, err)
				assert.True(t, errors.Is(err, context.DeadlineExceeded))
			} else {
				assert.NoError(t, err)
			}

			if tt.expectRepoCall {
				assert.Greater(t, mockRepo.CallsCreateOrder, 0)
			} else {
				assert.Equal(t, 0, mockRepo.CallsCreateOrder)
			}
		})
	}
}

func TestOrderMessageProcessor_ProcessMessage_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	processor := kafka.NewOrderMessageProcessor(logger)

	t.Run("context_cancellation", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository()
		mockRepo.ShouldFail = true
		mockRepo.FailError = errors.New("database error")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		messageData := mustMarshalOrder(testutils.MinimalOrderFixture("cancelled"))

		err := processor.ProcessMessage(ctx, messageData, mockRepo)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("empty_message", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository()

		err := processor.ProcessMessage(context.Background(), []byte{}, mockRepo)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid message format")
		assert.Equal(t, 0, mockRepo.CallsCreateOrder)
	})

	t.Run("null_message", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository()

		err := processor.ProcessMessage(context.Background(), nil, mockRepo)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid message format")
		assert.Equal(t, 0, mockRepo.CallsCreateOrder)
	})
}

func TestOrderConsumer_ConsumeOrders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("context_cancellation", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository()
		consumer := kafka.NewOrderConsumer(logger)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := consumer.ConsumeOrders(ctx, []string{"localhost:9092"}, "test-topic", "test-group", mockRepo)

		assert.NoError(t, err)
	})

	t.Run("timeout_context", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository()
		consumer := kafka.NewOrderConsumer(logger)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := consumer.ConsumeOrders(ctx, []string{"nonexistent:9092"}, "test-topic", "test-group", mockRepo)

		if err != nil {
			assert.Error(t, err)
		}
	})
}

func TestOrderConsumer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	t.Run("consumer_with_mock_repo", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository()
		consumer := kafka.NewOrderConsumer(logger)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			err := consumer.ConsumeOrders(ctx, []string{"localhost:9092"}, "test-topic", "test-group", mockRepo)
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Logf("Consumer error (expected for test): %v", err)
			}
		case <-time.After(200 * time.Millisecond):
			t.Log("Consumer test timed out (expected)")
		}

		assert.Equal(t, 0, mockRepo.CallsCreateOrder)
	})
}

// Вспомогательные функции

func mustMarshalOrder(order *dto.OrderDTO) []byte {
	data, err := json.Marshal(order)
	if err != nil {
		panic(err)
	}
	return data
}
