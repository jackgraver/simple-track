package services

import (
	"be-simpletracker/internal/clock"
	"context"
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
	day, err := dietrepo.DayMealPlanToday(ctx, offset)
	if err != nil {
		return models.DietDay{}, dietrepo.MealDayTotals{}, err
	}
	tot := dietrepo.CalculateTotals(ctx, day.ID)
	return day, tot, nil
}

func MealPlanWeek(ctx context.Context) ([]models.DietDay, error) {
	today := clock.Now()
	start := today.AddDate(0, 0, -3)
	end := today.AddDate(0, 0, 3)
	return dietrepo.DaysByDateRange(ctx, start, end)
}

func MealPlanMonth(ctx context.Context, offset int) (days []models.DietDay, startOfMonth, endOfMonth time.Time, month time.Month, err error) {
	today := clock.Now()
	target := today.AddDate(0, offset, 0)
	startOfMonth = time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
	endOfMonth = startOfMonth.AddDate(0, 1, -1)
	days, err = dietrepo.DaysByDateRange(ctx, startOfMonth, endOfMonth)
	if err != nil {
		return nil, startOfMonth, endOfMonth, 0, err
	}
	return days, startOfMonth, endOfMonth, target.Month(), nil
}

func MonthPlannedSummary(ctx context.Context, monthOffset int) ([]int, error) {
	today := clock.Now()
	target := today.AddDate(0, monthOffset, 0)
	loc := target.Location()
	startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	dim := endOfMonth.Day()
	byDate, err := dietrepo.CountUnloggedPlannedMealsPerCalendarDay(ctx, startOfMonth, endOfMonth)
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
	tot := dietrepo.CalculateTotals(ctx, day.ID)
	return day, tot, nil
}

func GoalsToday(ctx context.Context) (*models.Plan, error) {
	return dietrepo.GoalsToday(ctx)
}

func MealPlanDayByID(ctx context.Context, id int) (*models.DietDay, error) {
	return dietrepo.DayByID(ctx, id)
}

func CalculateTotals(ctx context.Context, dayID uint) dietrepo.MealDayTotals {
	return dietrepo.CalculateTotals(ctx, dayID)
}

func DayWithTotalsByID(ctx context.Context, id int) (DayWithTotals, error) {
	day, err := dietrepo.DayByID(ctx, id)
	if err != nil {
		return DayWithTotals{}, err
	}
	return DayWithTotals{Day: day, Totals: dietrepo.CalculateTotals(ctx, day.ID)}, nil
}

func AllMealDays(ctx context.Context) ([]models.DietDay, error) {
	return dietrepo.AllMealDays(ctx)
}

func AllFoods(ctx context.Context, excludeIDs []uint) ([]models.Food, error) {
	return dietrepo.FoodsAll(ctx, excludeIDs)
}

func AllFoodsForPicker(ctx context.Context, excludeIDs []uint) ([]models.FoodWithVariants, error) {
	return dietrepo.FoodsAllWithVariantSiblings(ctx, excludeIDs)
}

func CreateFood(ctx context.Context, food *models.Food, relatedFoodID *uint) (*models.Food, error) {
	food.ID = 0
	food.VariantGroupID = nil
	if err := dietrepo.FoodCreateWithOptionalRelated(ctx, food, relatedFoodID); err != nil {
		return nil, err
	}
	dietrepo.EnrichFoodVariants(ctx, food)
	return food, nil
}

func UpdateFood(ctx context.Context, id uint, in dietrepo.FoodFields) (*models.Food, error) {
	food, err := dietrepo.FoodUpdate(ctx, id, in)
	if err != nil {
		return nil, err
	}
	dietrepo.EnrichFoodVariants(ctx, food)
	return food, nil
}

func DeleteFood(ctx context.Context, id uint) error {
	return dietrepo.FoodDelete(ctx, id)
}

func AllCompositeFoods(ctx context.Context) ([]models.CompositeFood, error) {
	return dietrepo.CompositeFoodsAll(ctx)
}

func CreateCompositeFood(ctx context.Context, cf *models.CompositeFood) (uint, error) {
	return dietrepo.CompositeFoodCreate(ctx, cf)
}

func CompositeFoodByID(ctx context.Context, id uint) (*models.CompositeFood, error) {
	return dietrepo.CompositeFoodByID(ctx, id)
}

func AllMeals(ctx context.Context, excludeIDs []uint) ([]models.Meal, error) {
	return dietrepo.MealsAll(ctx, excludeIDs)
}

func MealByID(ctx context.Context, id uint) (*models.Meal, error) {
	return dietrepo.MealByID(ctx, id)
}

func FindMealPlanDay(ctx context.Context, date time.Time) (*models.DietDay, error) {
	return dietrepo.FindDayByDate(ctx, date)
}

func CreateDayMeal(ctx context.Context, dayMeal *models.DayLog) error {
	return dietrepo.CreateDayMeal(ctx, dayMeal)
}

func DayLogExistsForMeal(ctx context.Context, dayID uint, mealID uint) (bool, error) {
	return dietrepo.DayLogExists(ctx, dayID, mealID)
}

func CreateMeal(ctx context.Context, meal *models.Meal) (uint, error) {
	return dietrepo.MealCreate(ctx, meal)
}

func AllSavedMeals(ctx context.Context, excludeIDs []uint) ([]models.SavedMeal, error) {
	return dietrepo.SavedMealsAll(ctx, excludeIDs)
}

func CreateSavedMeal(ctx context.Context, sm *models.SavedMeal) (uint, error) {
	return dietrepo.SavedMealCreate(ctx, sm)
}

func SavedMealByID(ctx context.Context, id uint) (*models.SavedMeal, error) {
	return dietrepo.SavedMealByID(ctx, id)
}

func SavedMealPlannedUsageInfo(ctx context.Context, savedMealID uint) (SavedMealPlannedUsage, error) {
	var out SavedMealPlannedUsage
	count, err := dietrepo.CountUnloggedPlannedBySavedMealID(ctx, savedMealID)
	if err != nil {
		return out, err
	}
	out.ReferenceCount = count
	return out, nil
}

func DeleteSavedMeal(ctx context.Context, id uint) error {
	return dietrepo.DeleteSavedMeal(ctx, id)
}

func ReplaceSavedMeal(ctx context.Context, id uint, incoming *models.SavedMeal) error {
	return dietrepo.SavedMealReplace(ctx, id, incoming)
}

func SetPlannedMealLogged(ctx context.Context, dayID uint, mealID uint) error {
	return dietrepo.SetPlannedMealLogged(ctx, dayID, mealID)
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

func AddPlannedMealFromSavedMeal(ctx context.Context, offset int, savedMealID uint, clientID string) error {
	sm, err := dietrepo.SavedMealByID(ctx, savedMealID)
	if err != nil {
		return err
	}
	return dietrepo.AddPlannedMealFromSavedMeal(ctx, offset, savedMealID, mealFromSaved(sm), clientID)
}

// CopyPlannedMeals copies one day's planned meals onto other days. See dietrepo.CopyPlannedMeals.
func CopyPlannedMeals(ctx context.Context, fromOffset int, toOffsets []int) (int, error) {
	return dietrepo.CopyPlannedMeals(ctx, fromOffset, toOffsets)
}

func ReorderPlannedMeals(ctx context.Context, offset int, orderedIDs []uint) error {
	day, err := dietrepo.FindDayByDate(ctx, clock.ZerodTime(offset))
	if err != nil {
		return err
	}
	return dietrepo.PlannedMealReorder(ctx, day.ID, orderedIDs)
}

func DeletePlannedMeal(ctx context.Context, offset int, plannedMealID uint) error {
	day, err := dietrepo.FindDayByDate(ctx, clock.ZerodTime(offset))
	if err != nil {
		return err
	}
	return dietrepo.PlannedMealDelete(ctx, plannedMealID, day.ID)
}

// LogPlannedMeal logs a planned meal on the given day (idempotent). See dietrepo.LogPlannedMeal.
func LogPlannedMeal(ctx context.Context, dayID uint, mealID uint) error {
	return dietrepo.LogPlannedMeal(ctx, dayID, mealID)
}

func DeleteLoggedMeal(ctx context.Context, dayID uint, mealID uint) error {
	return dietrepo.DeleteLoggedMeal(ctx, dayID, mealID)
}

func UpdateDayLogMeal(ctx context.Context, dayID uint, oldMealID uint, newMealID uint) error {
	return dietrepo.UpdateDayLogMeal(ctx, dayID, oldMealID, newMealID)
}

func EditLoggedMeal(ctx context.Context, dayID uint, oldMealID uint, meal *models.Meal) error {
	return dietrepo.EditLoggedMeal(ctx, dayID, oldMealID, meal)
}

func QuickLogMeal(ctx context.Context, params dietrepo.QuickLogParams) (DayWithTotals, error) {
	dayID, err := dietrepo.QuickLogMeal(ctx, params)
	if err != nil {
		return DayWithTotals{}, err
	}
	return DayWithTotalsByID(ctx, int(dayID))
}

func GetAllPlans(ctx context.Context, params utils.QueryParams) (*utils.ListResult[models.Plan], error) {
	return dietrepo.PlansGetAll(ctx, params)
}

func UpdatePlanMacros(ctx context.Context, id uint, calories, protein, fiber, carbs, fat float32) (*models.Plan, error) {
	return dietrepo.UpdatePlanMacros(ctx, id, calories, protein, fiber, carbs, fat)
}

func FindMealPlanDayByRowIDOrToday(ctx context.Context, dayID uint) (*models.DietDay, error) {
	if dayID != 0 {
		return MealPlanDayByID(ctx, int(dayID))
	}
	return FindMealPlanDay(ctx, clock.ZerodTime(0))
}

func DayWithTotalsForDay(ctx context.Context, day *models.DietDay) DayWithTotals {
	return DayWithTotals{Day: day, Totals: CalculateTotals(ctx, day.ID)}
}

func ReloadDayWithTotals(ctx context.Context, day *models.DietDay) (DayWithTotals, error) {
	return DayWithTotalsByID(ctx, int(day.ID))
}
