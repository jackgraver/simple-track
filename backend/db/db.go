package db

import (
	"time"

	"gorm.io/gorm"

	"be-simpletracker/db/models"
)

// AddMeal creates a new meal with items
func AddMeal(db *gorm.DB, name string, items []models.MealItem) (*models.Meal, error) {
	meal := &models.Meal{
		Name:  name,
		Items: items,
	}
	
	if err := db.Create(meal).Error; err != nil {
		return nil, err
	}
	
	return meal, nil
}

func GetDailyGoals(db *gorm.DB, date time.Time) (*models.DayGoals, error) {
	var todayPlan models.MealPlanDay
	if err := db.Where("date = ?", date.Format("2006-01-02")).First(&todayPlan).Error; err != nil {
		return nil, err
	}

	var goals models.DayGoals
	if err := db.Where("meal_plan_day_id = ?", todayPlan.ID).First(&goals).Error; err != nil {
		return nil, err
	}

	return &goals, nil
}

// GetMeals retrieves all meals with their items
func GetMeals(db *gorm.DB) ([]models.Meal, error) {
	var meals []models.Meal
	if err := db.Preload("Items.Food").Find(&meals).Error; err != nil {
		return nil, err
	}
	return meals, nil
}

// AddFood creates a new food item
func AddFood(db *gorm.DB, name, unit string, calories, protein, fiber float32) (*models.Food, error) {
	food := &models.Food{
		Name:     name,
		Unit:     unit,
		Calories: calories,
		Protein:  protein,
		Fiber:    fiber,
	}
	
	if err := db.Create(food).Error; err != nil {
		return nil, err
	}
	
	return food, nil
}

// GetFoods retrieves all food items
func GetFoods(db *gorm.DB) ([]models.Food, error) {
	var foods []models.Food
	if err := db.Find(&foods).Error; err != nil {
		return nil, err
	}
	return foods, nil
}

// GetMealPlanDays retrieves all meal plan days
func GetMealPlanDays(db *gorm.DB) ([]models.MealPlanDay, error) {
	var days []models.MealPlanDay

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

// GetTodayMealPlan retrieves meal plan for today
func GetTodayMealPlan(db *gorm.DB, date time.Time) ([]models.MealPlanDay, error) {
	var days []models.MealPlanDay
	if err := db.Preload("Meals.Meal.Items.Food").Preload("Goals").Where("date = ?", date).Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

// GetMealNames retrieves all meal names
func GetMealNames(db *gorm.DB, limit int) ([]models.Meal, error) {
    var meals []models.Meal
    if err := db.Preload("Items.Food").
        Order("name").
        Limit(limit).
        Find(&meals).Error; err != nil {
        return nil, err
    }
    return meals, nil
}