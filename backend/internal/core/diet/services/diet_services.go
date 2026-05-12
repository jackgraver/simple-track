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

func CalculateTotals(db *gorm.DB, dayID uint) dietrepo.MealDayTotals {
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

// AllFoodsForPicker returns foods with sibling variants for the add-foods list.
func AllFoodsForPicker(db *gorm.DB, excludeIDs []uint) ([]models.FoodWithVariants, error) {
	return dietrepo.New(db).FoodsAllWithVariantSiblings(excludeIDs)
}

func CreateFood(db *gorm.DB, food *models.Food, relatedFoodID *uint) (*models.Food, error) {
	r := dietrepo.New(db)
	food.ID = 0
	food.VariantGroupID = nil
	if err := r.FoodCreateWithOptionalRelated(food, relatedFoodID); err != nil {
		return nil, err
	}
	r.EnrichFoodVariants(food)
	return food, nil
}

func AllCompositeFoods(db *gorm.DB) ([]models.CompositeFood, error) {
	return dietrepo.New(db).CompositeFoodsAll()
}

func CreateCompositeFood(db *gorm.DB, cf *models.CompositeFood) (uint, error) {
	return dietrepo.New(db).CompositeFoodCreate(cf)
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

func DayLogExistsForMeal(db *gorm.DB, dayID uint, mealID uint) (bool, error) {
	return dietrepo.New(db).DayLogExists(dayID, mealID)
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

type SavedMealPlannedUsage struct {
	ReferenceCount int64 `json:"reference_count"`
}

func SavedMealByID(db *gorm.DB, id uint) (*models.SavedMeal, error) {
	return dietrepo.New(db).SavedMealByID(id)
}

func SavedMealPlannedUsageInfo(db *gorm.DB, savedMealID uint) (SavedMealPlannedUsage, error) {
	r := dietrepo.New(db)
	var out SavedMealPlannedUsage
	count, err := r.CountUnloggedPlannedBySavedMealID(savedMealID)
	if err != nil {
		return out, err
	}
	out.ReferenceCount = count
	return out, nil
}

func DeleteSavedMeal(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		r := dietrepo.New(tx)
		if err := r.DeleteUnloggedPlannedBySavedMealID(id); err != nil {
			return err
		}
		return r.SavedMealDelete(id)
	})
}

func ReplaceSavedMeal(db *gorm.DB, id uint, incoming *models.SavedMeal) error {
	return dietrepo.New(db).SavedMealReplace(id, incoming)
}

func SetPlannedMealLogged(db *gorm.DB, dayID uint, mealID uint) error {
	return dietrepo.New(db).SetPlannedMealLogged(dayID, mealID)
}

func mealFromSaved(sm *models.SavedMeal) *models.Meal {
	m := &models.Meal{Name: sm.Name}
	for _, it := range sm.Items {
		m.Items = append(m.Items, models.MealItem{
			FoodID:            it.FoodID,
			Amount:            float32(it.Amount),
			GroupID:           it.GroupID,
			GroupLabel:        it.GroupLabel,
			CompositeFoodID:   it.CompositeFoodID,
		})
	}
	return m
}

// AddPlannedMealFromSavedMeal creates a new meal from a saved template and attaches it as planned for the calendar day at offset.
func AddPlannedMealFromSavedMeal(db *gorm.DB, offset int, savedMealID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		r := dietrepo.New(tx)
		day, err := r.FindDayByDate(utils.ZerodTime(offset))
		if err != nil {
			return err
		}
		sm, err := r.SavedMealByID(savedMealID)
		if err != nil {
			return err
		}
		meal := mealFromSaved(sm)
		mealID, err := r.MealCreate(meal)
		if err != nil {
			return err
		}
		displayOrder, err := r.NextPlannedMealDisplayOrder(day.ID)
		if err != nil {
			return err
		}
		sid := savedMealID
		return r.PlannedMealCreate(&models.PlannedMeal{
			DayID:        day.ID,
			MealID:       mealID,
			SavedMealID:  &sid,
			Logged:       false,
			DisplayOrder: displayOrder,
		})
	})
}

// ReorderPlannedMeals sets display_order for unlogged planned meals on the day at offset from orderedIDs (full list, in desired order).
func ReorderPlannedMeals(db *gorm.DB, offset int, orderedIDs []uint) error {
	r := dietrepo.New(db)
	day, err := r.FindDayByDate(utils.ZerodTime(offset))
	if err != nil {
		return err
	}
	return r.PlannedMealReorder(day.ID, orderedIDs)
}

// DeletePlannedMeal removes a planned meal row for the calendar day at offset.
func DeletePlannedMeal(db *gorm.DB, offset int, plannedMealID uint) error {
	r := dietrepo.New(db)
	day, err := r.FindDayByDate(utils.ZerodTime(offset))
	if err != nil {
		return err
	}
	return r.PlannedMealDelete(plannedMealID, day.ID)
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
