package dto

type ItemDTO struct {
	ChrtID      uint64 `json:"chrt_id" validate:"required,min=1"`
	TrackNumber string `json:"track_number" validate:"required,min=1,max=100"`
	Price       int    `json:"price" validate:"required,min=0"`
	RID         string `json:"rid" validate:"required,min=1,max=100"`
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Sale        int    `json:"sale" validate:"min=0,max=100"`
	Size        string `json:"size" validate:"required,min=1,max=10"`
	TotalPrice  int    `json:"total_price" validate:"required,min=0"`
	NmID        uint64 `json:"nm_id" validate:"required,min=1"`
	Brand       string `json:"brand" validate:"required,min=1,max=100"`
	Status      int    `json:"status" validate:"required,min=0"`
}
