package steps

import (
	"time"

	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

// StepLog is one step count per day per user (unique on (user_id, date), migration 0013).
type StepLog struct {
	gorm.Model
	database.Owner
	Date  time.Time `json:"date" gorm:"not null"`
	Steps int       `json:"steps" gorm:"not null"`
}

func (StepLog) TableName() string { return "step_logs" }
