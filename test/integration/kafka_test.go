package integration

import (
	"L0/internal/config"
	"fmt"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaConnection(t *testing.T) {
	cfg := config.MustLoad()
	brokers := cfg.Kafka.Brokers

	conn, err := kafka.Dial("tcp", brokers[0])

	require.NoError(t, err, "Failed to connect to Kafka")
	require.NotNil(t, conn, "Connection should not be nil")

	defer func(conn *kafka.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Failed to close connection: %v", err)
		}
	}(conn)

	partitions, err := conn.ReadPartitions()
	assert.NoError(t, err, "Failed to read partitions")
	assert.NotNil(t, partitions, "Partitions should not be nil")
}
