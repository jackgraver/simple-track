package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/utils"
)

type SavedMealPlannedUsage struct {
	ReferenceCount int64 `json:"reference_count"`
}

type DayWithTotals struct {
	Day    *models.DietDay
	Totals dietrepo.MealDayTotals
}

func MealPlanToday(ctx context.Context, offset int) (models.DietDay, dietrepo.MealDayTotals, error) {
	day, err := dietrepo.DayMealPlanToday(offset)
	if err != nil {
		return models.DietDay{}, dietrepo.MealDayTotals{}, err
	}
	tot := dietrepo.CalculateTotals(day.ID)
	return day, tot, nil
}

func MealPlanWeek(ctx context.Context) ([]models.DietDay, error) {
	today := time.Now()
	start := today.AddDate(0, 0, -3)
	end := today.AddDate(0, 0, 3)
	return dietrepo.DaysByDateRange(ctx, start, end)
}

func MealPlanMonth(ctx context.Context, offset int) (days []models.DietDay, startOfMonth, endOfMonth time.Time, month time.Month, err error) {
	today := time.Now()
	target := today.AddDate(0, offset, 0)
	startOfMonth = time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
	endOfMonth = startOfMonth.AddDate(0, 1, -1)
	days, err = dietrepo.DaysByDateRange(ctx, startOfMonth, endOfMonth)
	if err != nil {
		return nil, startOfMonth, endOfMonth, 0, err
	}
	return days, startOfMonth, endOfMonth, target.Month(), nil
}

func MonthPlannedSummary(_ context.Context, monthOffset int) ([]int, error) {
	today := time.Now()
	target := today.AddDate(0, monthOffset, 0)
	loc := target.Location()
	startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	dim := endOfMonth.Day()
	byDate, err := dietrepo.CountUnloggedPlannedMealsPerCalendarDay(startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}
	out := make([]int, dim)
	for day := 1; day <= dim; day++ {
		d := time.Date(target.Year(), target.Month(), day, 0, 0, 0, 0, loc)
		out[day-1] = byDate[d.Format("2006-01-02")]
	}
	return out, nil
}

func MealPlanDay(ctx context.Context, id uint) (models.DietDay, dietrepo.MealDayTotals, error) {
	day, err := dietrepo.DayByIDGeneric(ctx, id)
	if err != nil {
		return models.DietDay{}, dietrepo.MealDayTotals{}, err
	}
	tot := dietrepo.CalculateTotals(day.ID)
	return day, tot, nil
}

func GoalsToday() (*models.Plan, error) {
	return dietrepo.GoalsToday()
}

func MealPlanDayByID(id int) (*models.DietDay, error) {
	return dietrepo.DayByID(id)
}

func CalculateTotals(dayID uint) dietrepo.MealDayTotals {
	return dietrepo.CalculateTotals(dayID)
}

func DayWithTotalsByID(id int) (DayWithTotals, error) {
	day, err := dietrepo.DayByID(id)
	if err != nil {
		return DayWithTotals{}, err
	}
	return DayWithTotals{Day: day, Totals: dietrepo.CalculateTotals(day.ID)}, nil
}

func AllMealDays() ([]models.DietDay, error) {
	return dietrepo.AllMealDays()
}

func AllFoods(excludeIDs []uint) ([]models.Food, error) {
	return dietrepo.FoodsAll(excludeIDs)
}

func AllFoodsForPicker(excludeIDs []uint) ([]models.FoodWithVariants, error) {
	return dietrepo.FoodsAllWithVariantSiblings(excludeIDs)
}

func CreateFood(food *models.Food, relatedFoodID *uint) (*models.Food, error) {
	food.ID = 0
	food.VariantGroupID = nil
	if err := dietrepo.FoodCreateWithOptionalRelated(food, relatedFoodID); err != nil {
		return nil, err
	}
	dietrepo.EnrichFoodVariants(food)
	return food, nil
}

func AllCompositeFoods() ([]models.CompositeFood, error) {
	return dietrepo.CompositeFoodsAll()
}

func CreateCompositeFood(cf *models.CompositeFood) (uint, error) {
	return dietrepo.CompositeFoodCreate(cf)
}

func CompositeFoodByID(id uint) (*models.CompositeFood, error) {
	return dietrepo.CompositeFoodByID(id)
}

func AllMeals(excludeIDs []uint) ([]models.Meal, error) {
	return dietrepo.MealsAll(excludeIDs)
}

func MealByID(id uint) (*models.Meal, error) {
	return dietrepo.MealByID(id)
}

func FindMealPlanDay(date time.Time) (*models.DietDay, error) {
	return dietrepo.FindDayByDate(date)
}

func CreateDayMeal(dayMeal *models.DayLog) error {
	return dietrepo.CreateDayMeal(dayMeal)
}

func DayLogExistsForMeal(dayID uint, mealID uint) (bool, error) {
	return dietrepo.DayLogExists(dayID, mealID)
}

func CreateMeal(meal *models.Meal) (uint, error) {
	return dietrepo.MealCreate(meal)
}

func AllSavedMeals(excludeIDs []uint) ([]models.SavedMeal, error) {
	return dietrepo.SavedMealsAll(excludeIDs)
}

func CreateSavedMeal(sm *models.SavedMeal) (uint, error) {
	return dietrepo.SavedMealCreate(sm)
}

func SavedMealByID(id uint) (*models.SavedMeal, error) {
	return dietrepo.SavedMealByID(id)
}

func SavedMealPlannedUsageInfo(savedMealID uint) (SavedMealPlannedUsage, error) {
	var out SavedMealPlannedUsage
	count, err := dietrepo.CountUnloggedPlannedBySavedMealID(savedMealID)
	if err != nil {
		return out, err
	}
	out.ReferenceCount = count
	return out, nil
}

func DeleteSavedMeal(id uint) error {
	return dietrepo.DeleteSavedMeal(id)
}

func ReplaceSavedMeal(id uint, incoming *models.SavedMeal) error {
	return dietrepo.SavedMealReplace(id, incoming)
}

func SetPlannedMealLogged(dayID uint, mealID uint) error {
	return dietrepo.SetPlannedMealLogged(dayID, mealID)
}

func mealFromSaved(sm *models.SavedMeal) *models.Meal {
	m := &models.Meal{Name: sm.Name}
	for _, it := range sm.Items {
		m.Items = append(m.Items, models.MealItem{
			FoodID:          it.FoodID,
			Amount:          float32(it.Amount),
			GroupID:         it.GroupID,
			GroupLabel:      it.GroupLabel,
			CompositeFoodID: it.CompositeFoodID,
		})
	}
	return m
}

func AddPlannedMealFromSavedMeal(offset int, savedMealID uint) error {
	sm, err := dietrepo.SavedMealByID(savedMealID)
	if err != nil {
		return err
	}
	return dietrepo.AddPlannedMealFromSavedMeal(offset, savedMealID, mealFromSaved(sm))
}

func AddPlannedMealFromLabel(offset int, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("name is required")
	}
	return dietrepo.AddPlannedMealFromLabel(offset, &models.Meal{Name: trimmed})
}

func ReorderPlannedMeals(offset int, orderedIDs []uint) error {
	day, err := dietrepo.FindDayByDate(utils.ZerodTime(offset))
	if err != nil {
		return err
	}
	return dietrepo.PlannedMealReorder(day.ID, orderedIDs)
}

func DeletePlannedMeal(offset int, plannedMealID uint) error {
	day, err := dietrepo.FindDayByDate(utils.ZerodTime(offset))
	if err != nil {
		return err
	}
	return dietrepo.PlannedMealDelete(plannedMealID, day.ID)
}

func DeleteLoggedMeal(dayID uint, mealID uint) error {
	return dietrepo.DeleteLoggedMeal(dayID, mealID)
}

func UpdateDayLogMeal(dayID uint, oldMealID uint, newMealID uint) error {
	return dietrepo.UpdateDayLogMeal(dayID, oldMealID, newMealID)
}

func EditLoggedMeal(dayID uint, oldMealID uint, meal *models.Meal) error {
	return dietrepo.EditLoggedMeal(dayID, oldMealID, meal)
}

func QuickLogMeal(params dietrepo.QuickLogParams) (DayWithTotals, error) {
	dayID, err := dietrepo.QuickLogMeal(params)
	if err != nil {
		return DayWithTotals{}, err
	}
	return DayWithTotalsByID(int(dayID))
}

func GetAllPlans(ctx context.Context, params utils.QueryParams) (*utils.GetAllResult[models.Plan], error) {
	return dietrepo.PlansGetAll(ctx, params)
}

func UpdatePlanMacros(id uint, calories, protein, fiber, carbs, fat float32) (*models.Plan, error) {
	return dietrepo.UpdatePlanMacros(id, calories, protein, fiber, carbs, fat)
}

func FindMealPlanDayByRowIDOrToday(dayID uint) (*models.DietDay, error) {
	if dayID != 0 {
		return MealPlanDayByID(int(dayID))
	}
	return FindMealPlanDay(utils.ZerodTime(0))
}

func DayWithTotalsForDay(day *models.DietDay) DayWithTotals {
	return DayWithTotals{Day: day, Totals: CalculateTotals(day.ID)}
}

func ReloadDayWithTotals(day *models.DietDay) (DayWithTotals, error) {
	return DayWithTotalsByID(int(day.ID))
}
