package dto

type DeliveryDTO struct {
	Name    string `json:"name" validate:"required,min=1,max=100"`
	Phone   string `json:"phone" validate:"required,min=10,max=20"`
	Zip     string `json:"zip" validate:"required,min=5,max=10"`
	City    string `json:"city" validate:"required,min=1,max=100"`
	Address string `json:"address" validate:"required,min=1,max=200"`
	Region  string `json:"region" validate:"required,min=1,max=100"`
	Email   string `json:"email" validate:"required,email"`
}
