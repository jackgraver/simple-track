package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	gorm.Model
	Username  string `json:"username" gorm:"not null;uniqueIndex"`
	Password  string `json:"-" gorm:"not null"`
	Email     string `json:"email" gorm:"uniqueIndex"`
	BirthYear *int   `json:"birth_year"`
	IsAdmin   bool   `json:"is_admin" gorm:"not null;default:false"`
	// TokenVersion is copied into every JWT; bumping it revokes all of the user's sessions.
	// Starts at 1, so a token with no version (0) never matches.
	TokenVersion int64 `json:"-" gorm:"not null;default:1"`
	// DisabledAt set means the user can't sign in and their tokens are rejected.
	DisabledAt *time.Time `json:"disabled_at,omitempty"`
}

// InitialTokenVersion is the token_version a new user starts with (the column default).
const InitialTokenVersion int64 = 1

func (u User) TableName() string { return "users" }
