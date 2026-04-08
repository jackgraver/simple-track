package models

import (
	"time"

	"gorm.io/gorm"
)

// DietDay represents a day in the diet plan
type DietDay struct {
	gorm.Model
	Date         time.Time     `json:"date" gorm:"uniqueIndex;not null"`
	PlanID       uint          `json:"plan_id"`
	Plan         Plan          `gorm:"foreignKey:PlanID" json:"plan"`
	PlannedMeals []PlannedMeal `gorm:"foreignKey:DayID" json:"plannedMeals"`
	Logs         []DayLog      `gorm:"foreignKey:DayID" json:"loggedMeals"`
}

func (d DietDay) GetID() uint        { return d.ID }
func (d DietDay) TableName() string  { return "days" }
func (d DietDay) GetDate() time.Time { return d.Date }
func (d DietDay) Preloads() []string {
	return []string{"PlannedMeals.Meal.Items.Food", "Plan", "Logs.Meal.Items.Food"}
}

// DayLog represents a logged meal for a day
type DayLog struct {
	gorm.Model
	DayID  uint `json:"day_id" gorm:"not null"`
	MealID uint `json:"meal_id" gorm:"not null"`
	Meal   Meal `json:"meal"`
}

func (d DayLog) GetID() uint        { return d.ID }
func (d DayLog) TableName() string  { return "day_logs" }
func (d DayLog) Preloads() []string { return []string{"Meal.Items.Food"} }
