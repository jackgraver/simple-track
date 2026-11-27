package models

import (
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// Meal represents a logged meal instance (foods eaten in one sitting); day_logs reference this.
type Meal struct {
	gorm.Model
	Name  string     `json:"name" gorm:"not null"`
	Items []MealItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
	// ClientID is a client-generated uuid that makes creates idempotent (unique when set).
	ClientID *string `json:"client_id,omitempty" gorm:"type:uuid"`
}

func (m Meal) TableName() string { return "meals" }

// MealItem represents an item in a meal
type MealItem struct {
	gorm.Model
	MealID          uint    `json:"meal_id" gorm:"not null;index"`
	FoodID          uint    `json:"food_id" gorm:"not null;index"`
	Amount          float32 `json:"amount"`
	GroupID         string  `json:"group_id" gorm:"index"`
	GroupLabel      string  `json:"group_label"`
	CompositeFoodID *uint   `json:"composite_food_id"`
	Meal            Meal    `json:"meal"`
	Food            Food    `json:"food"`
}

func (m MealItem) TableName() string { return "meal_items" }

// SavedMeal is a reusable template (e.g. meals you eat often), not tied to a specific day log.
type SavedMeal struct {
	gorm.Model
	Name  string          `json:"name" gorm:"not null"`
	Items []SavedMealItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
}

func (s SavedMeal) TableName() string { return "saved_meals" }

// SavedMealItem represents an item in a saved meal
type SavedMealItem struct {
	gorm.Model
	SavedMealID     uint      `json:"saved_meal_id" gorm:"not null;index"`
	FoodID          uint      `json:"food_id" gorm:"not null;index"`
	Amount          float64   `json:"amount"`
	GroupID         string    `json:"group_id" gorm:"index"`
	GroupLabel      string    `json:"group_label"`
	CompositeFoodID *uint     `json:"composite_food_id"`
	SavedMeal       SavedMeal `json:"-"`
	Food            Food      `json:"food"`
}

func (s SavedMealItem) TableName() string { return "saved_meal_items" }

// Food represents a food item
type Food struct {
	gorm.Model
	Name           string  `json:"name" gorm:"not null;uniqueIndex"`
	ServingType    string  `json:"serving_type" gorm:"not null"`
	ServingAmount  float32 `json:"serving_amount" gorm:"not null"`
	Calories       float32 `json:"calories" gorm:"not null"`
	Protein        float32 `json:"protein"`
	Fiber          float32 `json:"fiber"`
	Carbs          float32 `json:"carbs"`
	Fat            float32 `json:"fat"`
	VariantGroupID *uint   `json:"variant_group_id" gorm:"index"`
	QuickEntry     bool    `json:"quick_entry" gorm:"default:false;index"`
	Variants       []Food  `json:"variants,omitempty" gorm:"-"`
}

func (f Food) TableName() string { return "foods" }

var quickLogFoodNameSuffix = regexp.MustCompile(`\s*\[ql-\d+\]\s*$`)

// NormalizeQuickLogFoodNameForResponse overwrites Name with the human-readable value for API
// responses only (in-memory; DB row unchanged). Applies to quick-entry rows or legacy names that
// end with our uniquifier suffix.
func NormalizeQuickLogFoodNameForResponse(f *Food) {
	if f == nil {
		return
	}
	if !f.QuickEntry && !quickLogFoodNameSuffix.MatchString(f.Name) {
		return
	}
	f.Name = strings.TrimRight(quickLogFoodNameSuffix.ReplaceAllString(f.Name, ""), " \t")
}

// FoodWithVariants is one food row for the food picker, including optional sibling variant foods.
type FoodWithVariants struct {
	Food
	Variants []Food `json:"variants,omitempty"`
}

// CompositeFood is a reusable recipe (multiple foods + amounts); logging expands to grouped meal items.
type CompositeFood struct {
	gorm.Model
	Name  string              `json:"name" gorm:"not null;uniqueIndex"`
	Items []CompositeFoodItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
}

func (c CompositeFood) TableName() string { return "composite_foods" }

// CompositeFoodItem is one ingredient line in a composite food.
type CompositeFoodItem struct {
	gorm.Model
	CompositeFoodID uint          `json:"composite_food_id" gorm:"not null;index"`
	FoodID          uint          `json:"food_id" gorm:"not null;index"`
	Amount          float32       `json:"amount"`
	CompositeFood   CompositeFood `json:"-"`
	Food            Food          `json:"food"`
}

func (c CompositeFoodItem) TableName() string { return "composite_food_items" }
