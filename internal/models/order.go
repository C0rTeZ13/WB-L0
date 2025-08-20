package models

import (
	"time"
)

type Order struct {
	ID                uint64 `gorm:"primaryKey;autoIncrement"`
	OrderUID          string `gorm:"type:text;unique;not null"`
	TrackNumber       string `gorm:"type:text;unique;not null"`
	Entry             string `gorm:"type:text;not null"`
	Locale            string `gorm:"type:text"`
	InternalSignature string `gorm:"type:text"`
	CustomerID        string `gorm:"type:text;not null"`
	DeliveryService   string `gorm:"type:text;not null"`
	ShardKey          string `gorm:"column:shardkey;type:text;not null"`
	SmID              uint64
	DateCreated       *time.Time
	OofShard          string `gorm:"type:text"`

	DeliveryID uint64
	PaymentID  uint64

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	Delivery Delivery `gorm:"constraint:OnDelete:SET NULL;"`
	Payment  Payment  `gorm:"constraint:OnDelete:SET NULL;"`
	Items    []Item   `gorm:"foreignKey:OrderID"`
}
