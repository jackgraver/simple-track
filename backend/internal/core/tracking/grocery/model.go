package grocery

import (
	"time"

	"gorm.io/gorm"
)

type GroceryItem struct {
	gorm.Model
	Name        string     `json:"name" gorm:"not null;index"`
	CompletedAt *time.Time `json:"completed_at"`
	// ClientID is a client-generated uuid that makes creates idempotent (unique when set).
	ClientID *string `json:"client_id,omitempty" gorm:"type:uuid"`
}

func (GroceryItem) TableName() string { return "grocery_items" }
