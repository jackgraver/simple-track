package models

import "time"

type MealPlanDay struct {
    ID    uint       `gorm:"primaryKey"`
    Date  time.Time  `gorm:"type:date" json:"date"`
    Meals []DayMeal     `gorm:"foreignKey:DayID" json:"meals"`
    Goals DayGoals   `gorm:"foreignKey:MealPlanDayID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"goals"`
    BaseModel
}

type DayGoals struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	MealPlanDayID uint       `gorm:"uniqueIndex" json:"meal_plan_day_id"`
	Calories      float32    `json:"calories"`
	Protein       float32    `json:"protein"`
	Fiber         float32    `json:"fiber"`
	BaseModel
}
