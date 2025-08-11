package integration

import (
	"L0/internal/config"
	"fmt"
	"github.com/segmentio/kafka-go"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestKafkaConnection(t *testing.T) {
	Convey("Given a running Kafka broker", t, func() {
		cfg := config.MustLoad()
		brokers := cfg.Kafka.Brokers

		Convey("When connecting to Kafka", func() {
			conn, err := kafka.Dial("tcp", brokers[0])

			Convey("Then the connection should succeed", func() {
				So(err, ShouldBeNil)
				So(conn, ShouldNotBeNil)
			})

			if err == nil {
				defer func(conn *kafka.Conn) {
					err := conn.Close()
					if err != nil {
						fmt.Printf("Failed to close connection")
					}
				}(conn)

				Convey("And Kafka should return partitions metadata", func() {
					partitions, err := conn.ReadPartitions()
					So(err, ShouldBeNil)

					So(partitions, ShouldNotBeNil)
				})
			}
		})
	})
}
