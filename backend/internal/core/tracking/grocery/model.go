package grocery

import (
	"time"

	"gorm.io/gorm"
)

type GroceryItem struct {
	gorm.Model
	Name        string     `json:"name" gorm:"not null;index"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (GroceryItem) TableName() string { return "grocery_items" }
