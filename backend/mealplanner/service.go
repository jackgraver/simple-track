package mealplanner

import (
	"time"

	"gorm.io/gorm"
)

// MealPlanToday returns today's meal plan day with meals and goals
func MealPlanToday(db *gorm.DB) ([]Day, error) {
    today := time.Now().Truncate(24 * time.Hour)
	var days []Day
	if err := db.Preload("Meals.Meal.Items.Food").Preload("Goals").Where("date = ?", today).Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

// MealPlanRange returns a simple 7-day window centered on today
func MealPlanRange(db *gorm.DB, today time.Time, start time.Time, end time.Time) ([]Day, error) {
    var days []Day

	if err := db.
		Preload("PlannedMeals.Meal.Items.Food").
		Preload("Plan").
        Preload("Logs.Meal.Items.Food").
		Where("date BETWEEN ? AND ?", start, end).
		Order("date").
		Find(&days).Error; err != nil {
		return nil, err
	}

	return days, nil
}

func MealPlanDayByID(db *gorm.DB, id int) (*Day, error) {
    var day Day

    if err := db.
        Preload("Meals.Meal.Items.Food").
        Preload("Goals").
		Preload("Meals").
        First(&day, id).Error; err != nil {
        return nil, err
    }

    return &day, nil
}

func AllMealDays(db *gorm.DB) ([]Day, error) {
    var days []Day

	if err := db.
		Find(&days).Error; err != nil {
		return nil, err
	}

	return days, nil
}

func GoalsToday(db *gorm.DB) (*Plan, error) {
    today := time.Now().Truncate(24 * time.Hour)

	var todayPlan Day
	if err := db.Where("date = ?", today.Format("2006-01-02")).First(&todayPlan).Error; err != nil {
		return nil, err
	}

	var goals Plan
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
    if err := db.Preload("Items.Food").Find(&meals).Distinct("name").Error; err != nil {
        return nil, err
    }
    return meals, nil
}

func FindMealPlanDay(db *gorm.DB, date time.Time) (*Day, error) {
    var day Day
    err := db.Where("date = ?", date).First(&day).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            // Explicitly return nil if no record exists
            return nil, nil
        }
        return nil, err
    }
    return &day, nil
}


func CreateDayMeal(db *gorm.DB, dayMeal *DayLog) error {
    if err := db.Create(&dayMeal).Error; err != nil {
        return err
    }
    return nil
}