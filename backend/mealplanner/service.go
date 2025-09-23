package mealplanner

import (
	"time"

	"gorm.io/gorm"
)

// MealPlanToday returns today's meal plan day with meals and goals
func MealPlanToday(db *gorm.DB) ([]MealPlanDay, error) {
    today := time.Now().Truncate(24 * time.Hour)
	var days []MealPlanDay
	if err := db.Preload("Meals.Meal.Items.Food").Preload("Goals").Where("date = ?", today).Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

// MealPlanWeek returns a simple 7-day window centered on today
func MealPlanWeek(db *gorm.DB) ([]MealPlanDay, error) {
    var days []MealPlanDay

	today := time.Now()
	start := today.AddDate(0, 0, -3) // 3 days before
	end := today.AddDate(0, 0, 3)    // 3 days after

	if err := db.
		Preload("Meals.Meal.Items.Food").
		Preload("Goals").
		Where("date BETWEEN ? AND ?", start, end).
		Order("date").
		Find(&days).Error; err != nil {
		return nil, err
	}

	return days, nil
}

func GoalsToday(db *gorm.DB) (*DayGoals, error) {
    today := time.Now().Truncate(24 * time.Hour)

	var todayPlan MealPlanDay
	if err := db.Where("date = ?", today.Format("2006-01-02")).First(&todayPlan).Error; err != nil {
		return nil, err
	}

	var goals DayGoals
	if err := db.Where("meal_plan_day_id = ?", todayPlan.ID).First(&goals).Error; err != nil {
		return nil, err
	}

	return &goals, nil
}

func AllFoods(db *gorm.DB) ([]Food, error) {
    var foods []Food
    if err := db.Find(&foods).Error; err != nil {
        return nil, err
    }
    return foods, nil
}

func AllMeals(db *gorm.DB) ([]Meal, error) {
    var meals []Meal
    if err := db.Find(&meals).Distinct("name").Error; err != nil {
        return nil, err
    }
    return meals, nil
}