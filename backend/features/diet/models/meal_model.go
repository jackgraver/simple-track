package models

import "gorm.io/gorm"

// Meal represents a meal
type Meal struct {
	gorm.Model
	Name  string     `json:"name" gorm:"not null"`
	Items []MealItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
}

func (m Meal) GetID() uint        { return m.ID }
func (m Meal) TableName() string  { return "meals" }
func (m Meal) Preloads() []string { return []string{"Items.Food"} }

// MealItem represents an item in a meal
type MealItem struct {
	gorm.Model
	MealID uint    `json:"meal_id" gorm:"not null;index"`
	FoodID uint    `json:"food_id" gorm:"not null;index"`
	Amount float32 `json:"amount"`
	Meal   Meal    `json:"meal"`
	Food   Food    `json:"food"`
}

func (m MealItem) GetID() uint        { return m.ID }
func (m MealItem) TableName() string  { return "meal_items" }
func (m MealItem) Preloads() []string { return []string{"Food"} }

// SavedMeal represents a saved meal
type SavedMeal struct {
	gorm.Model
	Name  string          `json:"name" gorm:"not null"`
	Items []SavedMealItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
}

func (s SavedMeal) GetID() uint        { return s.ID }
func (s SavedMeal) TableName() string  { return "saved_meals" }
func (s SavedMeal) Preloads() []string { return []string{"Items.Food"} }

// SavedMealItem represents an item in a saved meal
type SavedMealItem struct {
	gorm.Model
	SavedMealID uint    `json:"saved_meal_id" gorm:"not null;index"`
	FoodID      uint    `json:"food_id" gorm:"not null;index"`
	Amount      float64 `json:"amount"`
	SavedMeal   SavedMeal
	Food        Food
}

func (s SavedMealItem) GetID() uint       { return s.ID }
func (s SavedMealItem) TableName() string { return "saved_meal_items" }

// Food represents a food item
type Food struct {
	gorm.Model
	Name          string  `json:"name" gorm:"not null;uniqueIndex"`
	ServingType   string  `json:"serving_type" gorm:"not null"`
	ServingAmount float32 `json:"serving_amount" gorm:"not null"`
	Calories      float32 `json:"calories" gorm:"not null"`
	Protein       float32 `json:"protein"`
	Fiber         float32 `json:"fiber"`
	Carbs         float32 `json:"carbs"`
}

func (f Food) GetID() uint       { return f.ID }
func (f Food) TableName() string { return "foods" }