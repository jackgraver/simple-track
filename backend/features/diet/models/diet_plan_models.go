package models

import "gorm.io/gorm"

// Plan represents a diet plan
type Plan struct {
    gorm.Model
    Name string `json:"name"`
	Calories float32 `json:"calories"`
    Protein  float32 `json:"protein"`
    Fiber    float32 `json:"fiber"`
	Carbs    float32 `json:"carbs"`
}

func (p Plan) GetID() uint        { return p.ID }
func (p Plan) TableName() string  { return "plans" }
func (p Plan) Preloads() []string { return []string{} }

// PlannedMeal represents a meal that is planned for a day
type PlannedMeal struct {
	gorm.Model
	DayID  uint `json:"day_id" gorm:"not null"`
	Day    Day  `json:"day"`
	MealID uint `json:"meal_id" gorm:"not null"`
	Meal   Meal `json:"meal"`
	Logged bool `json:"logged"`
}

func (p PlannedMeal) GetID() uint        { return p.ID }
func (p PlannedMeal) TableName() string  { return "planned_meals" }
func (p PlannedMeal) Preloads() []string { return []string{"Meal.Items.Food"} }
