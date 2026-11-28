package profile

import (
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

// UserProfile is one row per user (unique live user_id, migration 0013), created on first save.
type UserProfile struct {
	gorm.Model
	database.Owner
	HeightIn      float64 `json:"height_in" gorm:"not null"`
	Age           int     `json:"age" gorm:"not null"`
	Sex           string  `json:"sex" gorm:"not null"`
	ActivityLevel string  `json:"activity_level" gorm:"not null"`
}

func (UserProfile) TableName() string { return "user_profiles" }
