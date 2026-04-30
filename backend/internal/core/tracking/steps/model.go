package steps

import (
	"time"

	"gorm.io/gorm"
)

type StepLog struct {
	gorm.Model
	Date  time.Time `json:"date" gorm:"uniqueIndex;not null"`
	Steps int       `json:"steps" gorm:"not null"`
}

func (StepLog) TableName() string { return "step_logs" }
