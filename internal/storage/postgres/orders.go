package postgres

import (
	"L0/ent"
	"L0/internal/kafka/dto"
	"context"
	"time"

	_ "entgo.io/ent/dialect/sql"
	_ "github.com/google/uuid"
)

func CreateOrder(ctx context.Context, client *ent.Client, o *dto.OrderDTO) (*ent.Order, error) {
	parsedDateCreated, err := time.Parse(time.RFC3339, o.DateCreated)

	if err != nil {
		return nil, err
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	d, err := tx.Delivery.Create().
		SetName(o.Delivery.Name).
		SetPhone(o.Delivery.Phone).
		SetZip(o.Delivery.Zip).
		SetCity(o.Delivery.City).
		SetAddress(o.Delivery.Address).
		SetRegion(o.Delivery.Region).
		SetEmail(o.Delivery.Email).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	p, err := tx.Payment.Create().
		SetTransaction(o.Payment.Transaction).
		SetRequestID(o.Payment.RequestID).
		SetCurrency(o.Payment.Currency).
		SetProvider(o.Payment.Provider).
		SetAmount(o.Payment.Amount).
		SetPaymentDt(o.Payment.PaymentDt).
		SetBank(o.Payment.Bank).
		SetDeliveryCost(o.Payment.DeliveryCost).
		SetGoodsTotal(o.Payment.GoodsTotal).
		SetCustomFee(o.Payment.CustomFee).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	var items []*ent.Item
	for _, it := range o.Items {
		itCopy := it
		itemEntity, err := tx.Item.Create().
			SetChrtID(itCopy.ChrtID).
			SetTrackNumber(itCopy.TrackNumber).
			SetPrice(itCopy.Price).
			SetRid(itCopy.RID).
			SetName(itCopy.Name).
			SetSale(itCopy.Sale).
			SetSize(itCopy.Size).
			SetTotalPrice(itCopy.TotalPrice).
			SetNmID(itCopy.NmID).
			SetBrand(itCopy.Brand).
			SetStatus(itCopy.Status).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		items = append(items, itemEntity)
	}

	ordBuilder := tx.Order.Create().
		SetOrderUID(o.OrderUID).
		SetTrackNumber(o.TrackNumber).
		SetEntry(o.Entry).
		SetLocale(o.Locale).
		SetInternalSignature(o.InternalSignature).
		SetCustomerID(o.CustomerID).
		SetDeliveryService(o.DeliveryService).
		SetShardkey(o.Shardkey).
		SetSmID(o.SmID).
		SetDateCreated(parsedDateCreated).
		SetOofShard(o.OofShard).
		SetDelivery(d).
		SetPayment(p).
		AddItems(items...)

	orderEntity, err := ordBuilder.Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return orderEntity, nil
}

func GetAllOrders(ctx context.Context, client *ent.Client) ([]dto.OrderDTO, error) {
	ordersEnt, err := client.Order.
		Query().
		WithDelivery().
		WithPayment().
		WithItems().
		All(ctx)
	if err != nil {
		return nil, err
	}

	var orders []dto.OrderDTO

	for _, o := range ordersEnt {
		order := dto.OrderDTO{
			OrderUID:          o.OrderUID,
			TrackNumber:       o.TrackNumber,
			Entry:             o.Entry,
			Locale:            *o.Locale,
			InternalSignature: *o.InternalSignature,
			CustomerID:        o.CustomerID,
			DeliveryService:   o.DeliveryService,
			Shardkey:          o.Shardkey,
			SmID:              *o.SmID,
			DateCreated:       o.DateCreated.Format(time.RFC3339),
			OofShard:          *o.OofShard,
		}

		if o.Edges.Delivery != nil {
			order.Delivery = dto.DeliveryDTO{
				Name:    o.Edges.Delivery.Name,
				Phone:   o.Edges.Delivery.Phone,
				Zip:     o.Edges.Delivery.Zip,
				City:    o.Edges.Delivery.City,
				Address: o.Edges.Delivery.Address,
				Region:  o.Edges.Delivery.Region,
				Email:   o.Edges.Delivery.Email,
			}
		}

		if o.Edges.Payment != nil {
			order.Payment = dto.PaymentDTO{
				Transaction:  o.Edges.Payment.Transaction,
				RequestID:    *o.Edges.Payment.RequestID,
				Currency:     o.Edges.Payment.Currency,
				Provider:     o.Edges.Payment.Provider,
				Amount:       o.Edges.Payment.Amount,
				PaymentDt:    o.Edges.Payment.PaymentDt,
				Bank:         o.Edges.Payment.Bank,
				DeliveryCost: o.Edges.Payment.DeliveryCost,
				GoodsTotal:   o.Edges.Payment.GoodsTotal,
				CustomFee:    o.Edges.Payment.CustomFee,
			}
		}

		for _, itemEnt := range o.Edges.Items {
			item := dto.ItemDTO{
				ChrtID:      itemEnt.ChrtID,
				TrackNumber: itemEnt.TrackNumber,
				Price:       itemEnt.Price,
				RID:         itemEnt.Rid,
				Name:        itemEnt.Name,
				Sale:        itemEnt.Sale,
				Size:        *itemEnt.Size,
				TotalPrice:  itemEnt.TotalPrice,
				NmID:        itemEnt.NmID,
				Brand:       *itemEnt.Brand,
				Status:      itemEnt.Status,
			}
			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	return orders, nil
}
