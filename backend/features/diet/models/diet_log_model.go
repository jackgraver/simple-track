package models

import (
	"time"

	"gorm.io/gorm"
)

// Day represents a day in the diet plan
type Day struct {
	gorm.Model
	Date         time.Time     `json:"date"`
	PlanID       uint          `json:"plan_id"`
	Plan         Plan          `gorm:"foreignKey:PlanID" json:"plan"`
	PlannedMeals []PlannedMeal `json:"plannedMeals"`
	Logs         []DayLog      `json:"loggedMeals"`
}

func (d Day) GetID() uint        { return d.ID }
func (d Day) TableName() string  { return "days" }
func (d Day) GetDate() time.Time { return d.Date }
func (d Day) Preloads() []string {
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
