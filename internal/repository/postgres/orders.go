package postgres

import (
	"L0/internal/kafka/dto"
	"L0/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

func CreateOrder(ctx context.Context, db *gorm.DB, o *dto.OrderDTO) (*models.Order, error) {
	parsedDateCreated, err := time.Parse(time.RFC3339, o.DateCreated)
	if err != nil {
		return nil, err
	}

	var order models.Order

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delivery
		delivery := models.Delivery{
			Name:    o.Delivery.Name,
			Phone:   o.Delivery.Phone,
			Zip:     o.Delivery.Zip,
			City:    o.Delivery.City,
			Address: o.Delivery.Address,
			Region:  o.Delivery.Region,
			Email:   o.Delivery.Email,
		}
		if err := tx.Create(&delivery).Error; err != nil {
			return err
		}

		// Payment
		payment := models.Payment{
			Transaction:  o.Payment.Transaction,
			RequestID:    o.Payment.RequestID,
			Currency:     o.Payment.Currency,
			Provider:     o.Payment.Provider,
			Amount:       o.Payment.Amount,
			PaymentDT:    o.Payment.PaymentDt,
			Bank:         o.Payment.Bank,
			DeliveryCost: o.Payment.DeliveryCost,
			GoodsTotal:   o.Payment.GoodsTotal,
			CustomFee:    o.Payment.CustomFee,
		}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		// Items
		var items []models.Item
		for _, it := range o.Items {
			item := models.Item{
				ChrtID:      it.ChrtID,
				TrackNumber: it.TrackNumber,
				Price:       it.Price,
				RID:         it.RID,
				Name:        it.Name,
				Sale:        it.Sale,
				Size:        it.Size,
				TotalPrice:  it.TotalPrice,
				NmID:        it.NmID,
				Brand:       it.Brand,
				Status:      it.Status,
			}
			items = append(items, item)
		}

		// Order
		order = models.Order{
			OrderUID:          o.OrderUID,
			TrackNumber:       o.TrackNumber,
			Entry:             o.Entry,
			Locale:            o.Locale,
			InternalSignature: o.InternalSignature,
			CustomerID:        o.CustomerID,
			DeliveryService:   o.DeliveryService,
			ShardKey:          o.Shardkey,
			SmID:              o.SmID,
			DateCreated:       &parsedDateCreated,
			OofShard:          o.OofShard,
			DeliveryID:        delivery.ID,
			PaymentID:         payment.ID,
			Items:             items,
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func GetAllOrders(ctx context.Context, db *gorm.DB) ([]dto.OrderDTO, error) {
	var orders []models.Order
	if err := db.WithContext(ctx).
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Find(&orders).Error; err != nil {
		return nil, err
	}

	var result []dto.OrderDTO
	for _, o := range orders {
		order := dto.OrderDTO{
			OrderUID:          o.OrderUID,
			TrackNumber:       o.TrackNumber,
			Entry:             o.Entry,
			Locale:            o.Locale,
			InternalSignature: o.InternalSignature,
			CustomerID:        o.CustomerID,
			DeliveryService:   o.DeliveryService,
			Shardkey:          o.ShardKey,
			SmID:              o.SmID,
			DateCreated:       o.DateCreated.Format(time.RFC3339),
			OofShard:          o.OofShard,
		}

		if o.Delivery.ID != 0 {
			order.Delivery = dto.DeliveryDTO{
				Name:    o.Delivery.Name,
				Phone:   o.Delivery.Phone,
				Zip:     o.Delivery.Zip,
				City:    o.Delivery.City,
				Address: o.Delivery.Address,
				Region:  o.Delivery.Region,
				Email:   o.Delivery.Email,
			}
		}

		if o.Payment.ID != 0 {
			order.Payment = dto.PaymentDTO{
				Transaction:  o.Payment.Transaction,
				RequestID:    o.Payment.RequestID,
				Currency:     o.Payment.Currency,
				Provider:     o.Payment.Provider,
				Amount:       o.Payment.Amount,
				PaymentDt:    o.Payment.PaymentDT,
				Bank:         o.Payment.Bank,
				DeliveryCost: o.Payment.DeliveryCost,
				GoodsTotal:   o.Payment.GoodsTotal,
				CustomFee:    o.Payment.CustomFee,
			}
		}

		for _, item := range o.Items {
			order.Items = append(order.Items, dto.ItemDTO{
				ChrtID:      item.ChrtID,
				TrackNumber: item.TrackNumber,
				Price:       item.Price,
				RID:         item.RID,
				Name:        item.Name,
				Sale:        item.Sale,
				Size:        item.Size,
				TotalPrice:  item.TotalPrice,
				NmID:        item.NmID,
				Brand:       item.Brand,
				Status:      item.Status,
			})
		}

		result = append(result, order)
	}

	return result, nil
}

func GetOrderByUID(ctx context.Context, db *gorm.DB, orderUID string) (*dto.OrderDTO, error) {
	var order models.Order
	if err := db.WithContext(ctx).
		Preload("Delivery").
		Preload("Payment").
		Preload("Items").
		Where("order_uid = ?", orderUID).
		First(&order).Error; err != nil {
		return nil, err
	}

	result := dto.OrderDTO{
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.ShardKey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated.Format(time.RFC3339),
		OofShard:          order.OofShard,
	}

	if order.Delivery.ID != 0 {
		result.Delivery = dto.DeliveryDTO{
			Name:    order.Delivery.Name,
			Phone:   order.Delivery.Phone,
			Zip:     order.Delivery.Zip,
			City:    order.Delivery.City,
			Address: order.Delivery.Address,
			Region:  order.Delivery.Region,
			Email:   order.Delivery.Email,
		}
	}

	if order.Payment.ID != 0 {
		result.Payment = dto.PaymentDTO{
			Transaction:  order.Payment.Transaction,
			RequestID:    order.Payment.RequestID,
			Currency:     order.Payment.Currency,
			Provider:     order.Payment.Provider,
			Amount:       order.Payment.Amount,
			PaymentDt:    order.Payment.PaymentDT,
			Bank:         order.Payment.Bank,
			DeliveryCost: order.Payment.DeliveryCost,
			GoodsTotal:   order.Payment.GoodsTotal,
			CustomFee:    order.Payment.CustomFee,
		}
	}

	for _, item := range order.Items {
		result.Items = append(result.Items, dto.ItemDTO{
			ChrtID:      item.ChrtID,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			RID:         item.RID,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmID:        item.NmID,
			Brand:       item.Brand,
			Status:      item.Status,
		})
	}

	return &result, nil
}
