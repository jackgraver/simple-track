package services

import (
	"context"
	"time"

	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/utils"

	"gorm.io/gorm"
)

// MealPlanToday returns today's meal plan day with meals and goals.
func MealPlanToday(db *gorm.DB, offset int) (models.DietDay, error) {
	return dietrepo.New(db).DayMealPlanToday(offset)
}

func MealPlanDayByID(db *gorm.DB, id int) (*models.DietDay, error) {
	return dietrepo.New(db).DayByID(id)
}

func CalculateTotals(db *gorm.DB, dayID uint) (float32, float32, float32, float32) {
	return dietrepo.New(db).CalculateTotals(dayID)
}

func AllMealDays(db *gorm.DB) ([]models.DietDay, error) {
	return dietrepo.New(db).AllMealDays()
}

func GoalsToday(db *gorm.DB) (*models.Plan, error) {
	return dietrepo.New(db).GoalsToday()
}

func AllFoods(db *gorm.DB, excludeIDs []uint) ([]models.Food, error) {
	return dietrepo.New(db).FoodsAll(excludeIDs)
}

func CreateFood(db *gorm.DB, food *models.Food) (*models.Food, error) {
	r := dietrepo.New(db)
	if err := r.FoodCreate(food); err != nil {
		return nil, err
	}
	return food, nil
}

func AllMeals(db *gorm.DB, excludeIDs []uint) ([]models.Meal, error) {
	return dietrepo.New(db).MealsAll(excludeIDs)
}

func MealByID(db *gorm.DB, id uint) (*models.Meal, error) {
	return dietrepo.New(db).MealByID(id)
}

func FindMealPlanDay(db *gorm.DB, date time.Time) (*models.DietDay, error) {
	return dietrepo.New(db).FindDayByDate(date)
}

func CreateDayMeal(db *gorm.DB, dayMeal *models.DayLog) error {
	return dietrepo.New(db).CreateDayMeal(dayMeal)
}

func CreateMeal(db *gorm.DB, meal *models.Meal) (uint, error) {
	return dietrepo.New(db).MealCreate(meal)
}

func AllSavedMeals(db *gorm.DB, excludeIDs []uint) ([]models.SavedMeal, error) {
	return dietrepo.New(db).SavedMealsAll(excludeIDs)
}

func CreateSavedMeal(db *gorm.DB, sm *models.SavedMeal) (uint, error) {
	return dietrepo.New(db).SavedMealCreate(sm)
}

func SetPlannedMealLogged(db *gorm.DB, dayID uint, mealID uint) error {
	return dietrepo.New(db).SetPlannedMealLogged(dayID, mealID)
}

func DeleteLoggedMeal(db *gorm.DB, dayID uint, mealID uint) error {
	return dietrepo.New(db).DeleteLoggedMeal(dayID, mealID)
}

func UpdateDayLogMeal(db *gorm.DB, dayID uint, oldMealID uint, newMealID uint) error {
	return dietrepo.New(db).UpdateDayLogMeal(dayID, oldMealID, newMealID)
}

// GetAllPlans lists plans with optional pagination and filters from parsed query params.
func GetAllPlans(ctx context.Context, db *gorm.DB, params utils.QueryParams) (*utils.GetAllResult[models.Plan], error) {
	return dietrepo.New(db).PlansGetAll(ctx, params)
}
