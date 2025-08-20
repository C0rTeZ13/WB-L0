package testutils

import (
	"L0/internal/kafka/dto"
	"time"
)

// OrderFixture создает тестовый заказ с валидными данными
func OrderFixture() *dto.OrderDTO {
	return &dto.OrderDTO{
		OrderUID:          "b563feb7b2b84b6test",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          "1",
		Delivery: dto.DeliveryDTO{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: dto.PaymentDTO{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []dto.ItemDTO{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}
}

// MinimalOrderFixture создает минимальный заказ для тестов
func MinimalOrderFixture(orderUID string) *dto.OrderDTO {
	trackNumber := "TEST_TRACK_" + orderUID
	return &dto.OrderDTO{
		OrderUID:          orderUID,
		TrackNumber:       trackNumber,
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test_customer",
		DeliveryService:   "test_delivery",
		Shardkey:          "1",
		SmID:              123,
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          "1",
		Delivery: dto.DeliveryDTO{
			Name:    "Test User",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "Test City",
			Address: "123 Test Street",
			Region:  "Test Region",
			Email:   "test@example.com",
		},
		Payment: dto.PaymentDTO{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "test_provider",
			Amount:       100,
			PaymentDt:    1637907727,
			Bank:         "test_bank",
			DeliveryCost: 0,
			GoodsTotal:   100,
			CustomFee:    0,
		},
		Items: []dto.ItemDTO{{
			ChrtID:      123,
			TrackNumber: trackNumber,
			Price:       100,
			RID:         "test_rid_" + orderUID,
			Name:        "Test Item",
			Sale:        0,
			Size:        "M",
			TotalPrice:  100,
			NmID:        456,
			Brand:       "Test Brand",
			Status:      1,
		}},
	}
}
