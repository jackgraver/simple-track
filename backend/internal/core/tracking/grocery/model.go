package grocery

import (
	"time"

	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

// GroceryItem is owned (migration 0013): each user has their own list.
type GroceryItem struct {
	gorm.Model
	database.Owner
	Name        string     `json:"name" gorm:"not null;index"`
	CompletedAt *time.Time `json:"completed_at"`
	// ClientID is a client-generated uuid that makes creates idempotent (unique per user when set).
	ClientID *string `json:"client_id,omitempty" gorm:"type:uuid"`
}

func (GroceryItem) TableName() string { return "grocery_items" }
