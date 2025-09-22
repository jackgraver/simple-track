package models

import (
	"time"

	"gorm.io/gorm"
)

type MealPlanDay struct {
	gorm.Model
	ID    uint
	Date  time.Time
	Meals []DayMeal
	Goals DayGoals `gorm:"foreignKey:MealPlanDayID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type DayGoals struct {
	gorm.Model
	ID            uint
	MealPlanDayID uint `gorm:"uniqueIndex"`
	Calories      float32
	Protein       float32
	Fiber         float32
}

type DayMeal struct {
	gorm.Model
	ID     uint
	DayID  uint `gorm:"not null"`
	MealID uint `gorm:"not null"`
	Meal   Meal
	Status string `gorm:"not null;default:'expected'"`
}

type Meal struct {
	gorm.Model
	ID    uint
	Name  string `gorm:"not null"`
	Items []MealItem
}

type MealItem struct {
	gorm.Model
	ID     uint
	MealID uint `gorm:"not null"`
	FoodID uint `gorm:"not null"`
	Food   Food
	Amount float64
}

type Food struct {
	gorm.Model
	ID       uint
	Name     string  `gorm:"not null"`
	Unit     string  `gorm:"not null"`
	Calories float32 `gorm:"not null"`
	Protein  float32
	Fiber    float32
}
