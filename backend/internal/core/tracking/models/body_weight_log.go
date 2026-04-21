package models

import (
	"time"

	"gorm.io/gorm"
)

type BodyWeightLog struct {
	gorm.Model
	Date      time.Time `json:"date" gorm:"uniqueIndex;not null"`
	WeightLbs float64   `json:"weight_lbs" gorm:"not null"`
}

func (BodyWeightLog) TableName() string { return "body_weight_logs" }
