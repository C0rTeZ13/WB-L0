package models

import (
	"time"
)

type Payment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Transaction  string    `gorm:"size:255;not null"`
	RequestID    string    `gorm:"size:255"`
	Currency     string    `gorm:"size:10;not null"`
	Provider     string    `gorm:"size:100;not null"`
	Amount       int       `gorm:"not null"`
	PaymentDT    uint64    `gorm:"not null"`
	Bank         string    `gorm:"size:100;not null"`
	DeliveryCost int       `gorm:"not null"`
	GoodsTotal   int       `gorm:"not null"`
	CustomFee    int       `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	Orders []Order `gorm:"foreignKey:PaymentID"`
}
