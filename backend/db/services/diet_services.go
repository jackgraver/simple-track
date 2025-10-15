package services

import (
	"be-simpletracker/db/models"
	"time"

	"gorm.io/gorm"
)

// MealPlanToday returns today's meal plan day with meals and goals
func MealPlanToday(db *gorm.DB) (models.Day, error) {
    now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	var days models.Day
	if err := db.Preload("PlannedMeals.Meal.Items.Food").
            Preload("Plan").
            Preload("Logs.Meal.Items.Food").
            Where("date >= ? AND date < ?", start, end).
            First(&days).Error; err != nil {
		return models.Day{}, err
	}
	return days, nil
}

func MealPlanDayByID(db *gorm.DB, id int) (*models.Day, error) {
    var day models.Day

    if err := db.
        Preload("PlannedMeals.Meal.Items.Food").
        Preload("Plan").
        Preload("Logs.Meal.Items.Food").
        First(&day, id).Error; err != nil {
        return nil, err
    }

    return &day, nil
}

func AllMealDays(db *gorm.DB) ([]models.Day, error) {
    var days []models.Day

	if err := db.
		Find(&days).Error; err != nil {
		return nil, err
	}

	return days, nil
}

func GoalsToday(db *gorm.DB) (*models.Plan, error) {
    today := time.Now().Truncate(24 * time.Hour)

	var todayPlan models.Day
	if err := db.Where("date = ?", today.Format("2006-01-02")).First(&todayPlan).Error; err != nil {
		return nil, err
	}

	var goals models.Plan
	if err := db.Where("meal_plan_day_id = ?", todayPlan.ID).First(&goals).Error; err != nil {
		return nil, err
	}

	return &goals, nil
}

func AllFoods(db *gorm.DB) ([]models.Food, error) {
    var foods []models.Food
    if err := db.Find(&foods).Error; err != nil {
        return nil, err
    }
    return foods, nil
}

func AllMeals(db *gorm.DB) ([]models.Meal, error) {
    var meals []models.Meal
    if err := db.Preload("Items.Food").Find(&meals).Distinct("name").Error; err != nil {
        return nil, err
    }
    return meals, nil
}

func FindMealPlanDay(db *gorm.DB, date time.Time) (*models.Day, error) {
    var day models.Day
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


func CreateDayMeal(db *gorm.DB, dayMeal *models.DayLog) error {
    if err := db.Create(&dayMeal).Error; err != nil {
        return err
    }
    return nil
}