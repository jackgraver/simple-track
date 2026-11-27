package models

import (
	"time"

	"gorm.io/gorm"
)

// Plan represents a diet plan
type Plan struct {
	gorm.Model
	Name          string     `json:"name"`
	Calories      float32    `json:"calories"`
	Protein       float32    `json:"protein"`
	Fiber         float32    `json:"fiber"`
	Carbs         float32    `json:"carbs"`
	Fat           float32    `json:"fat"`
	EffectiveFrom *time.Time `json:"effective_from,omitempty" gorm:"index"`
}

func (p Plan) TableName() string { return "plans" }

// PlannedMeal represents a meal that is planned for a day
type PlannedMeal struct {
	gorm.Model
	DayID        uint    `json:"day_id" gorm:"not null"`
	DietDay      DietDay `gorm:"foreignKey:DayID" json:"day"`
	MealID       uint    `json:"meal_id" gorm:"not null"`
	Meal         Meal    `json:"meal"`
	SavedMealID  *uint   `json:"saved_meal_id,omitempty" gorm:"index"`
	Logged       bool    `json:"logged"`
	DisplayOrder int     `json:"display_order" gorm:"not null;default:0"`
	// ClientID is a client-generated uuid that makes creates idempotent (unique when set).
	ClientID *string `json:"client_id,omitempty" gorm:"type:uuid"`
}

func (p PlannedMeal) TableName() string { return "planned_meals" }
