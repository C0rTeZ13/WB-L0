package models

import (
	"time"
)

type Item struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	ChrtID      uint64    `gorm:"not null"`
	TrackNumber string    `gorm:"size:255;not null"`
	Price       int       `gorm:"not null"`
	RID         string    `gorm:"column:rid;size:255;not null"`
	Name        string    `gorm:"size:255;not null"`
	Sale        int       `gorm:"not null"`
	Size        string    `gorm:"size:50"`
	TotalPrice  int       `gorm:"not null"`
	NmID        uint64    `gorm:"not null"`
	Brand       string    `gorm:"size:255"`
	Status      int       `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	OrderID uint64 `gorm:"index"`
	Order   Order  `gorm:"constraint:OnDelete:CASCADE;"`
}
