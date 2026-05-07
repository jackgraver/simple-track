package profile

import "gorm.io/gorm"

type UserProfile struct {
	gorm.Model
	HeightIn      float64 `json:"height_in" gorm:"not null"`
	Age           int     `json:"age" gorm:"not null"`
	Sex           string  `json:"sex" gorm:"not null"`
	ActivityLevel string  `json:"activity_level" gorm:"not null"`
}

func (UserProfile) TableName() string { return "user_profiles" }
