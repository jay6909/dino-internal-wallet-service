package repository

import "github.com/google/uuid"

type CurrencyType struct {
	ID   uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	BaseTimeStamps
}
