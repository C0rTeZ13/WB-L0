package dto

type OrderDTO struct {
	OrderUID          string      `json:"order_uid"`
	TrackNumber       string      `json:"track_number"`
	Entry             string      `json:"entry"`
	Delivery          DeliveryDTO `json:"delivery"`
	Payment           PaymentDTO  `json:"payment"`
	Items             []ItemDTO   `json:"items"`
	Locale            string      `json:"locale"`
	InternalSignature string      `json:"internal_signature"`
	CustomerID        string      `json:"customer_id"`
	DeliveryService   string      `json:"delivery_service"`
	Shardkey          string      `json:"shardkey"`
	SmID              int         `json:"sm_id"`
	DateCreated       string      `json:"date_created"`
	OofShard          string      `json:"oof_shard"`
}
