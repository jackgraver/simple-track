package mealplanner

import (
    "time"

    "be-simpletracker/db"
    "be-simpletracker/db/models"

    "gorm.io/gorm"
)

// Today returns today's meal plan day with meals and goals
func Today(database *gorm.DB) ([]models.MealPlanDay, error) {
    today := time.Now().Truncate(24 * time.Hour)
    return db.GetTodayMealPlan(database, today)
}

// Week returns a simple 7-day window centered on today
func Week(database *gorm.DB) ([]models.MealPlanDay, error) {
    return db.GetMealPlanDays(database)
}

// AllFoods returns all foods
func AllFoods(database *gorm.DB) ([]models.Food, error) {
    return db.GetFoods(database)
}

// AllMeals returns all meals including items and food details
func AllMeals(database *gorm.DB) ([]models.Meal, error) {
    return db.GetMeals(database)
}