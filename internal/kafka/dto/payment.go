package dto

type PaymentDTO struct {
	Transaction  string `json:"transaction" validate:"required,min=1,max=100"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency" validate:"required,min=3,max=3,oneof=USD EUR RUB"`
	Provider     string `json:"provider" validate:"required,min=1,max=50"`
	Amount       int    `json:"amount" validate:"required,min=0"`
	PaymentDt    uint64 `json:"payment_dt" validate:"required,min=1"`
	Bank         string `json:"bank" validate:"required,min=1,max=50"`
	DeliveryCost int    `json:"delivery_cost" validate:"min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"min=0"`
	CustomFee    int    `json:"custom_fee" validate:"min=0"`
}
