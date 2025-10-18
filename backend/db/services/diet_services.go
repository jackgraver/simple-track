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
	if err := db.
            Preload("PlannedMeals", "logged = ?", false).
            Preload("PlannedMeals.Meal.Items.Food").
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
            Preload("PlannedMeals", "logged = ?", false).
            Preload("PlannedMeals.Meal.Items.Food").
            Preload("Plan").
            Preload("Logs.Meal.Items.Food").
            First(&day, id).Error; err != nil {
        return nil, err
    }

    return &day, nil
}

func CalculateTotals(db *gorm.DB, dayID uint) (float32, float32, float32) {
    var totals struct {
        TotalCalories float32 `json:"total_calories"`
        TotalProtein  float32 `json:"total_protein"`
        TotalFiber    float32 `json:"total_fiber"`
    }

    db.Raw(`
        SELECT
            SUM(f.calories * mi.amount) AS total_calories,
            SUM(f.protein  * mi.amount) AS total_protein,
            SUM(f.fiber    * mi.amount) AS total_fiber
        FROM day_logs dl
        JOIN meals m       ON dl.meal_id = m.id
        JOIN meal_items mi ON mi.meal_id = m.id
        JOIN foods f       ON f.id = mi.food_id
        WHERE dl.day_id = ?
    `, dayID).Scan(&totals)
    
    return totals.TotalCalories, totals.TotalProtein, totals.TotalFiber
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

func MealByID(db *gorm.DB, id uint) (*models.Meal, error) {
    var meal models.Meal
    if err := db.Preload("Items.Food").First(&meal, id).Error; err != nil {
        return nil, err
    }
    return &meal, nil
}

func FindMealPlanDay(db *gorm.DB, date time.Time) (*models.Day, error) {
    var day models.Day
    err := db.Where("date = ?", date).First(&day).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
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

func CreateMeal(db *gorm.DB, meal *models.Meal) (uint, error) {
    for i := range meal.Items {
        meal.Items[i].ID = 0 
    }

    if err := db.Create(meal).Error; err != nil {
        return 0, err
    }
    return meal.ID, nil
}

func SetPlannedMealLogged(db *gorm.DB, dayID uint, mealID uint) error {
    var meal models.PlannedMeal
    if err := db.Where("day_id = ? AND meal_id = ?", dayID, mealID).First(&meal).Error; err != nil {
        return err
    }
    meal.Logged = true
    return db.Save(&meal).Error
}