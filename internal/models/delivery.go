package models

import (
	"time"
)

type Delivery struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:255;not null"`
	Phone     string    `gorm:"size:50;not null"`
	Zip       string    `gorm:"size:20;not null"`
	City      string    `gorm:"size:100;not null"`
	Address   string    `gorm:"type:text;not null"`
	Region    string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	Orders []Order `gorm:"foreignKey:DeliveryID"`
}
