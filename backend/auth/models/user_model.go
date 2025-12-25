package models

import "gorm.io/gorm"

// User represents a user account
type User struct {
	gorm.Model
	Username string `json:"username" gorm:"not null;uniqueIndex"`
	Password string `json:"-" gorm:"not null"`
	Email    string `json:"email" gorm:"uniqueIndex"`
}

func (u User) GetID() uint       { return u.ID }
func (u User) TableName() string { return "users" }

