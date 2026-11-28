package models

import (
	"time"

	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

// DietDay represents a day in the diet plan
type DietDay struct {
	gorm.Model
	database.Owner
	Date         time.Time     `json:"date" gorm:"not null"`
	PlanID       uint          `json:"plan_id"`
	Plan         Plan          `gorm:"foreignKey:PlanID" json:"plan"`
	PlannedMeals []PlannedMeal `gorm:"foreignKey:DayID" json:"plannedMeals"`
	Logs         []DayLog      `gorm:"foreignKey:DayID" json:"loggedMeals"`
}

func (d DietDay) TableName() string { return "days" }

// DayLog represents a logged meal for a day
type DayLog struct {
	gorm.Model
	database.Owner
	DayID  uint `json:"day_id" gorm:"not null"`
	MealID uint `json:"meal_id" gorm:"not null"`
	Meal   Meal `json:"meal"`
}

func (d DayLog) TableName() string { return "day_logs" }
