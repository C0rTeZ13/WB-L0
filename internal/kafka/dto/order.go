package dto

type OrderDTO struct {
	OrderUID          string      `json:"order_uid" validate:"required,min=1,max=100"`
	TrackNumber       string      `json:"track_number" validate:"required,min=1,max=100"`
	Entry             string      `json:"entry" validate:"required,min=1,max=50"`
	Delivery          DeliveryDTO `json:"delivery" validate:"required"`
	Payment           PaymentDTO  `json:"payment" validate:"required"`
	Items             []ItemDTO   `json:"items" validate:"required,min=1,dive"`
	Locale            string      `json:"locale" validate:"required,min=2,max=5"`
	InternalSignature string      `json:"internal_signature"`
	CustomerID        string      `json:"customer_id" validate:"required,min=1,max=100"`
	DeliveryService   string      `json:"delivery_service" validate:"required,min=1,max=50"`
	Shardkey          string      `json:"shardkey" validate:"required,min=1,max=20"`
	SmID              uint64      `json:"sm_id" validate:"required,min=1"`
	DateCreated       string      `json:"date_created" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	OofShard          string      `json:"oof_shard" validate:"required,min=1,max=10"`
}
