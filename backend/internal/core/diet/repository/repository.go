package dietrepo

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"be-simpletracker/internal/core/diet/models"
	"be-simpletracker/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func EnrichFoodVariants(ctx context.Context, f *models.Food) {
	if f == nil {
		return
	}
	models.NormalizeQuickLogFoodNameForResponse(f)
	if f.VariantGroupID == nil {
		return
	}
	var sibs []models.Food
	// Best effort: a failed lookup leaves the food without variants.
	if err := database.Optional(ctx, func() error {
		return database.From(ctx).Where("variant_group_id = ? AND id != ?", *f.VariantGroupID, f.ID).Order("name ASC").Find(&sibs).Error
	}); err != nil {
		return
	}
	for i := range sibs {
		models.NormalizeQuickLogFoodNameForResponse(&sibs[i])
	}
	f.Variants = sibs
}

func enrichMealFoodVariants(ctx context.Context, m *models.Meal) {
	if m == nil {
		return
	}
	for i := range m.Items {
		if m.Items[i].Food.ID != 0 {
			EnrichFoodVariants(ctx, &m.Items[i].Food)
		}
	}
}

func enrichSavedMealFoodVariants(ctx context.Context, sm *models.SavedMeal) {
	if sm == nil {
		return
	}
	for i := range sm.Items {
		if sm.Items[i].Food.ID != 0 {
			EnrichFoodVariants(ctx, &sm.Items[i].Food)
		}
	}
}

func enrichDietDayFoodVariants(ctx context.Context, d *models.DietDay) {
	if d == nil {
		return
	}
	for i := range d.PlannedMeals {
		if d.PlannedMeals[i].Meal.ID != 0 {
			enrichMealFoodVariants(ctx, &d.PlannedMeals[i].Meal)
		}
	}
	for i := range d.Logs {
		if d.Logs[i].Meal.ID != 0 {
			enrichMealFoodVariants(ctx, &d.Logs[i].Meal)
		}
	}
}

func FoodsAll(ctx context.Context, excludeIDs []uint) ([]models.Food, error) {
	var foods []models.Food
	query := database.From(ctx).Model(&models.Food{}).Where("quick_entry = ? OR quick_entry IS NULL", false).Order("name ASC")
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Find(&foods).Error; err != nil {
		return nil, err
	}
	return foods, nil
}

func FoodCreate(ctx context.Context, food *models.Food) error {
	return database.From(ctx).Create(food).Error
}

// FoodCreateWithOptionalRelated inserts a food and, if set, links it to a related food's variant group.
// ErrFoodNameTaken is returned when a food with the same name already exists
// (unique index idx_foods_name, which also covers soft-deleted foods).
var ErrFoodNameTaken = errors.New("a food with this name already exists")

func FoodCreateWithOptionalRelated(ctx context.Context, food *models.Food, relatedFoodID *uint) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		food.ID = 0
		if err := tx.Create(food).Error; err != nil {
			if isUniqueConstraintError(err) {
				return ErrFoodNameTaken
			}
			return err
		}
		if relatedFoodID == nil || *relatedFoodID == 0 {
			return nil
		}
		return linkNewFoodToRelatedInTx(tx, food, *relatedFoodID)
	})
}

func linkNewFoodToRelatedInTx(tx *gorm.DB, food *models.Food, relatedFoodID uint) error {
	if food.ID == 0 {
		return errors.New("food must be persisted before linking")
	}
	if food.ID == relatedFoodID {
		return nil
	}
	var related models.Food
	if err := tx.First(&related, relatedFoodID).Error; err != nil {
		return err
	}
	var gid uint
	if related.VariantGroupID != nil {
		gid = *related.VariantGroupID
	} else {
		var m uint
		if err := tx.Model(&models.Food{}).Select("COALESCE(MAX(variant_group_id), 0)").Scan(&m).Error; err != nil {
			return err
		}
		gid = m + 1
		if err := tx.Model(&related).Update("variant_group_id", gid).Error; err != nil {
			return err
		}
	}
	if err := tx.Model(food).Update("variant_group_id", gid).Error; err != nil {
		return err
	}
	food.VariantGroupID = &gid
	return nil
}

func foodsByVariantGroup(ctx context.Context) (map[uint][]models.Food, error) {
	var list []models.Food
	if err := database.From(ctx).Model(&models.Food{}).Where("variant_group_id IS NOT NULL").Order("name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	m := make(map[uint][]models.Food)
	for _, f := range list {
		if f.VariantGroupID == nil {
			continue
		}
		g := *f.VariantGroupID
		m[g] = append(m[g], f)
	}
	return m, nil
}

// FoodsAllWithVariantSiblings returns all foods (with excludes) and attaches sibling rows per variant group.
func FoodsAllWithVariantSiblings(ctx context.Context, excludeIDs []uint) ([]models.FoodWithVariants, error) {
	foods, err := FoodsAll(ctx, excludeIDs)
	if err != nil {
		return nil, err
	}
	byGroup, err := foodsByVariantGroup(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]models.FoodWithVariants, 0, len(foods))
	for _, f := range foods {
		row := models.FoodWithVariants{Food: f}
		models.NormalizeQuickLogFoodNameForResponse(&row.Food)
		if f.VariantGroupID != nil {
			for _, s := range byGroup[*f.VariantGroupID] {
				if s.ID != f.ID {
					sibling := s
					models.NormalizeQuickLogFoodNameForResponse(&sibling)
					row.Variants = append(row.Variants, sibling)
				}
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func CompositeFoodsAll(ctx context.Context) ([]models.CompositeFood, error) {
	var list []models.CompositeFood
	if err := database.From(ctx).Preload("Items.Food").Order("name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		for j := range list[i].Items {
			models.NormalizeQuickLogFoodNameForResponse(&list[i].Items[j].Food)
		}
	}
	return list, nil
}

func CompositeFoodCreate(ctx context.Context, cf *models.CompositeFood) (uint, error) {
	cf.ID = 0
	for i := range cf.Items {
		cf.Items[i].ID = 0
	}
	if err := database.From(ctx).Create(cf).Error; err != nil {
		return 0, err
	}
	return cf.ID, nil
}

func MealsAll(ctx context.Context, excludeIDs []uint) ([]models.Meal, error) {
	var meals []models.Meal
	query := database.From(ctx).Model(&models.Meal{}).Order("name ASC")
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Preload("Items.Food").Find(&meals).Distinct("name").Error; err != nil {
		return nil, err
	}
	return meals, nil
}

func SavedMealsAll(ctx context.Context, excludeIDs []uint) ([]models.SavedMeal, error) {
	var meals []models.SavedMeal
	query := database.From(ctx).Model(&models.SavedMeal{}).Order("name ASC")
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Preload("Items.Food").Find(&meals).Distinct("name").Error; err != nil {
		return nil, err
	}
	for i := range meals {
		if meals[i].Items != nil {
			enrichSavedMealFoodVariants(ctx, &meals[i])
		}
	}
	return meals, nil
}

func SavedMealCreate(ctx context.Context, sm *models.SavedMeal) (uint, error) {
	for i := range sm.Items {
		sm.Items[i].ID = 0
	}
	if err := database.From(ctx).Create(sm).Error; err != nil {
		return 0, err
	}
	return sm.ID, nil
}

func SavedMealByID(ctx context.Context, id uint) (*models.SavedMeal, error) {
	var sm models.SavedMeal
	if err := database.From(ctx).Preload("Items.Food").First(&sm, id).Error; err != nil {
		return nil, err
	}
	enrichSavedMealFoodVariants(ctx, &sm)
	return &sm, nil
}

func SavedMealDelete(ctx context.Context, id uint) error {
	var sm models.SavedMeal
	if err := database.From(ctx).First(&sm, id).Error; err != nil {
		return err
	}
	return database.From(ctx).Delete(&sm).Error
}

func CountUnloggedPlannedBySavedMealID(ctx context.Context, savedMealID uint) (int64, error) {
	var count int64
	err := database.From(ctx).Model(&models.PlannedMeal{}).
		Where("saved_meal_id = ? AND logged = ?", savedMealID, false).
		Count(&count).Error
	return count, err
}

func DeleteUnloggedPlannedBySavedMealID(ctx context.Context, savedMealID uint) error {
	return database.From(ctx).Where("saved_meal_id = ? AND logged = ?", savedMealID, false).Delete(&models.PlannedMeal{}).Error
}

// SavedMealReplace updates the template name and replaces all items in one transaction.
func SavedMealReplace(ctx context.Context, id uint, incoming *models.SavedMeal) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.SavedMeal
		if err := tx.First(&existing, id).Error; err != nil {
			return err
		}
		if err := tx.Where("saved_meal_id = ?", id).Delete(&models.SavedMealItem{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&existing).Update("name", incoming.Name).Error; err != nil {
			return err
		}
		for i := range incoming.Items {
			row := incoming.Items[i]
			row.ID = 0
			row.SavedMealID = id
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func PlannedMealCreate(ctx context.Context, pm *models.PlannedMeal) error {
	pm.ID = 0
	return database.From(ctx).Create(pm).Error
}

func PlannedMealDelete(ctx context.Context, plannedMealID uint, dayID uint) error {
	var pm models.PlannedMeal
	if err := database.From(ctx).Where("id = ? AND day_id = ?", plannedMealID, dayID).First(&pm).Error; err != nil {
		return err
	}
	return database.From(ctx).Delete(&pm).Error
}

func NextPlannedMealDisplayOrder(ctx context.Context, dayID uint) (int, error) {
	var next int
	err := database.From(ctx).Model(&models.PlannedMeal{}).
		Where("day_id = ? AND logged = ?", dayID, false).
		Select("COALESCE(MAX(display_order), -1) + 1").
		Scan(&next).Error
	return next, err
}

func PlannedMealReorder(ctx context.Context, dayID uint, orderedIDs []uint) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.PlannedMeal{}).
			Where("day_id = ? AND logged = ?", dayID, false).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if len(orderedIDs) == 0 {
				return nil
			}
			return fmt.Errorf("no planned meals to reorder")
		}
		if int(count) != len(orderedIDs) {
			return fmt.Errorf("planned meal count mismatch: expected %d ids, got %d", count, len(orderedIDs))
		}
		seen := make(map[uint]struct{}, len(orderedIDs))
		for _, id := range orderedIDs {
			if _, dup := seen[id]; dup {
				return fmt.Errorf("duplicate planned meal id in reorder list")
			}
			seen[id] = struct{}{}
			var pm models.PlannedMeal
			if err := tx.Where("id = ? AND day_id = ? AND logged = ?", id, dayID, false).First(&pm).Error; err != nil {
				return err
			}
		}
		for i, id := range orderedIDs {
			if err := tx.Model(&models.PlannedMeal{}).
				Where("id = ? AND day_id = ?", id, dayID).
				Update("display_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func MealByID(ctx context.Context, id uint) (*models.Meal, error) {
	var meal models.Meal
	if err := database.From(ctx).Preload("Items.Food").First(&meal, id).Error; err != nil {
		return nil, err
	}
	enrichMealFoodVariants(ctx, &meal)
	return &meal, nil
}

func MealCreate(ctx context.Context, meal *models.Meal) (uint, error) {
	for i := range meal.Items {
		meal.Items[i].ID = 0
	}
	if err := database.From(ctx).Create(meal).Error; err != nil {
		return 0, err
	}
	return meal.ID, nil
}

func calendarDayRange(t time.Time) (start, end time.Time) {
	loc := t.Location()
	start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	end = start.Add(24 * time.Hour)
	return start, end
}

func createDefaultPlan(db *gorm.DB) (models.Plan, error) {
	p := models.Plan{
		Name:     "Default",
		Calories: 2000,
		Protein:  150,
		Fiber:    30,
		Carbs:    200,
		Fat:      65,
	}
	if err := db.Create(&p).Error; err != nil {
		return models.Plan{}, err
	}
	return p, nil
}

func planForDate(db *gorm.DB, t time.Time) (models.Plan, error) {
	start, _ := calendarDayRange(t)
	var plan models.Plan
	err := db.
		Where("effective_from IS NOT NULL AND effective_from <= ?", start).
		Order("effective_from DESC, id DESC").
		First(&plan).Error
	if err == nil {
		return plan, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Plan{}, err
	}

	if err := db.Order("id ASC").First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return createDefaultPlan(db)
		}
		return models.Plan{}, err
	}
	return plan, nil
}

// findOrCreateDietDayForCalendarDate returns a day row for the wall-clock calendar day of t (location from t).
func findOrCreateDietDayForCalendarDate(ctx context.Context, t time.Time) (models.DietDay, error) {
	start, end := calendarDayRange(t)
	var day models.DietDay
	err := database.From(ctx).Where("date >= ? AND date < ?", start, end).First(&day).Error
	if err == nil {
		return day, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.DietDay{}, err
	}

	plan, err := planForDate(database.From(ctx), t)
	if err != nil {
		return models.DietDay{}, err
	}

	loc := t.Location()
	atMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	day = models.DietDay{
		Date:   atMidnight,
		PlanID: plan.ID,
	}
	if err := database.CreateInSavepoint(database.From(ctx), &day); err != nil {
		if isUniqueConstraintError(err) {
			if err := database.From(ctx).Where("date >= ? AND date < ?", start, end).First(&day).Error; err != nil {
				return models.DietDay{}, err
			}
			return day, nil
		}
		return models.DietDay{}, err
	}
	return day, nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate key") ||
		strings.Contains(s, "unique constraint") ||
		strings.Contains(s, "unique constraint failed")
}

func loadDietDayWithPreloads(ctx context.Context, id uint) (models.DietDay, error) {
	var day models.DietDay
	if err := database.From(ctx).
		Preload("PlannedMeals", func(db *gorm.DB) *gorm.DB {
			return db.Where("logged = ?", false).Order("display_order ASC, id ASC")
		}).
		Preload("PlannedMeals.Meal.Items.Food").
		Preload("Plan").
		Preload("Logs.Meal.Items.Food").
		First(&day, id).Error; err != nil {
		return models.DietDay{}, err
	}
	enrichDietDayFoodVariants(ctx, &day)
	return day, nil
}

func DayMealPlanToday(ctx context.Context, offset int) (models.DietDay, error) {
	d, err := findOrCreateDietDayForCalendarDate(ctx, clock.ZerodTime(offset))
	if err != nil {
		return models.DietDay{}, err
	}
	return loadDietDayWithPreloads(ctx, d.ID)
}

// CountUnloggedPlannedMealsPerCalendarDay returns YYYY-MM-DD (loc) → count for that calendar day.
func CountUnloggedPlannedMealsPerCalendarDay(ctx context.Context, start, end time.Time) (map[string]int, error) {
	out := make(map[string]int)
	rows, err := database.From(ctx).Model(&models.PlannedMeal{}).
		Joins("JOIN days ON days.id = planned_meals.day_id").
		Where("planned_meals.logged = ?", false).
		Where("days.date >= ? AND days.date <= ?", start, end).
		Where("planned_meals.deleted_at IS NULL").
		Where("days.deleted_at IS NULL").
		Select("days.date").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		loc := d.Location()
		k := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc).Format("2006-01-02")
		out[k]++
	}
	return out, rows.Err()
}

func DayByID(ctx context.Context, id int) (*models.DietDay, error) {
	day, err := loadDietDayWithPreloads(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return &day, nil
}

// preloadDietDay loads a day's plan, planned meals and logged meals with their foods.
func preloadDietDay(db *gorm.DB) *gorm.DB {
	return db.Preload("PlannedMeals.Meal.Items.Food").Preload("Plan").Preload("Logs.Meal.Items.Food")
}

func DaysByDateRange(ctx context.Context, start, end time.Time) ([]models.DietDay, error) {
	var days []models.DietDay
	err := database.From(ctx).WithContext(ctx).Scopes(preloadDietDay).
		Where("date >= ? AND date <= ?", start, end).
		Order("date ASC").
		Find(&days).Error
	return days, err
}

func DayByIDGeneric(ctx context.Context, id uint) (models.DietDay, error) {
	var day models.DietDay
	err := database.From(ctx).WithContext(ctx).Scopes(preloadDietDay).First(&day, id).Error
	return day, err
}

// MealDayTotals aggregates logged macros for one diet day from day_logs.
type MealDayTotals struct {
	Calories float32
	Protein  float32
	Fiber    float32
	Carbs    float32
	Fat      float32
}

func CalculateTotals(ctx context.Context, dayID uint) MealDayTotals {
	var totals MealDayTotals
	// Best effort, as before: on error the totals stay zero.
	_ = database.Optional(ctx, func() error {
		return database.From(ctx).Raw(`
        SELECT
            SUM(f.calories * mi.amount) AS calories,
            SUM(f.protein  * mi.amount) AS protein,
            SUM(f.fiber    * mi.amount) AS fiber,
            SUM(f.carbs    * mi.amount) AS carbs,
            SUM(COALESCE(f.fat, 0) * mi.amount) AS fat
        FROM day_logs dl
        JOIN meals m       ON dl.meal_id = m.id
        JOIN meal_items mi ON mi.meal_id = m.id
        JOIN foods f       ON f.id = mi.food_id
        WHERE dl.day_id = ?
        AND dl.deleted_at IS NULL
    `, dayID).Scan(&totals).Error
	})
	return totals
}

func AllMealDays(ctx context.Context) ([]models.DietDay, error) {
	var days []models.DietDay
	if err := database.From(ctx).Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

func GoalsToday(ctx context.Context) (*models.Plan, error) {
	todayDay, err := findOrCreateDietDayForCalendarDate(ctx, clock.ZerodTime(0))
	if err != nil {
		return nil, err
	}
	var plan models.Plan
	if err := database.From(ctx).First(&plan, todayDay.PlanID).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func FindDayByDate(ctx context.Context, date time.Time) (*models.DietDay, error) {
	day, err := findOrCreateDietDayForCalendarDate(ctx, date)
	if err != nil {
		return nil, err
	}
	return &day, nil
}

func CreateDayMeal(ctx context.Context, dayMeal *models.DayLog) error {
	return database.From(ctx).Create(dayMeal).Error
}

func DayLogExists(ctx context.Context, dayID uint, mealID uint) (bool, error) {
	var count int64
	err := database.From(ctx).Model(&models.DayLog{}).
		Where("day_id = ? AND meal_id = ?", dayID, mealID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func SetPlannedMealLogged(ctx context.Context, dayID uint, mealID uint) error {
	var pm models.PlannedMeal
	if err := database.From(ctx).Where("day_id = ? AND meal_id = ?", dayID, mealID).First(&pm).Error; err != nil {
		return err
	}
	pm.Logged = true
	return database.From(ctx).Save(&pm).Error
}

// DeleteLoggedMeal removes a logged meal from a day. It is idempotent: a log that is already
// gone is success, so a repeated (or retried) unlog does not fail.
//
// It does not reset planned_meals.logged: whether an unlogged planned meal returns to
// "planned" is an open owner question (optimistic-ui README, Q1).
func DeleteLoggedMeal(ctx context.Context, dayID uint, mealID uint) error {
	return database.From(ctx).Where("day_id = ? AND meal_id = ?", dayID, mealID).Delete(&models.DayLog{}).Error
}

// LogPlannedMeal marks the day's planned meal as logged and ensures exactly one day log exists
// for it. Repeating the call is a no-op. Returns gorm.ErrRecordNotFound when the day has no such
// planned meal.
func LogPlannedMeal(ctx context.Context, dayID uint, mealID uint) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		// FOR UPDATE serializes concurrent calls for the same planned meal even when it is
		// already logged=true (after an unlog, which keeps the flag), so the day_log count below
		// sees the other call's insert instead of both inserting and one hitting
		// uq_day_logs_day_meal.
		var pm models.PlannedMeal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("day_id = ? AND meal_id = ?", dayID, mealID).First(&pm).Error; err != nil {
			return err
		}
		if !pm.Logged {
			if err := tx.Model(&pm).Update("logged", true).Error; err != nil {
				return err
			}
		}
		var count int64
		if err := tx.Model(&models.DayLog{}).
			Where("day_id = ? AND meal_id = ?", dayID, mealID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		return tx.Create(&models.DayLog{DayID: dayID, MealID: mealID}).Error
	})
}

func UpdateDayLogMeal(ctx context.Context, dayID uint, oldMealID uint, newMealID uint) error {
	var log models.DayLog
	if err := database.From(ctx).Where("day_id = ? AND meal_id = ?", dayID, oldMealID).First(&log).Error; err != nil {
		return err
	}
	log.MealID = newMealID
	return database.From(ctx).Save(&log).Error
}

func PlansGetAll(ctx context.Context, params utils.QueryParams) (*utils.ListResult[models.Plan], error) {
	return utils.List[models.Plan](database.From(ctx).WithContext(ctx), params, "id", true)
}

func CompositeFoodByID(ctx context.Context, id uint) (*models.CompositeFood, error) {
	var cf models.CompositeFood
	if err := database.From(ctx).Preload("Items.Food").First(&cf, id).Error; err != nil {
		return nil, err
	}
	for i := range cf.Items {
		models.NormalizeQuickLogFoodNameForResponse(&cf.Items[i].Food)
	}
	return &cf, nil
}

func UpdatePlanMacros(ctx context.Context, id uint, calories, protein, fiber, carbs, fat float32) (*models.Plan, error) {
	effectiveStart, _ := calendarDayRange(clock.ZerodTime(0))
	var plan models.Plan

	if err := database.From(ctx).Transaction(func(tx *gorm.DB) error {
		var basePlan models.Plan
		if err := tx.First(&basePlan, id).Error; err != nil {
			return err
		}

		plan = models.Plan{
			Name:          basePlan.Name,
			Calories:      calories,
			Protein:       protein,
			Fiber:         fiber,
			Carbs:         carbs,
			Fat:           fat,
			EffectiveFrom: &effectiveStart,
		}
		if err := tx.Create(&plan).Error; err != nil {
			return err
		}

		return tx.Model(&models.DietDay{}).
			Where("date >= ?", effectiveStart).
			Update("plan_id", plan.ID).Error
	}); err != nil {
		return nil, err
	}
	return &plan, nil
}

func DeleteSavedMeal(ctx context.Context, id uint) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("saved_meal_id = ? AND logged = ?", id, false).Delete(&models.PlannedMeal{}).Error; err != nil {
			return err
		}
		var sm models.SavedMeal
		if err := tx.First(&sm, id).Error; err != nil {
			return err
		}
		return tx.Delete(&sm).Error
	})
}

// lockClientID serializes creates that carry the same client_id (a retry racing the original
// request, or a double submit) with a transaction-scoped advisory lock on Postgres. The caller's
// "does a row with this client_id exist?" check then runs after the other transaction committed,
// so the replay returns cleanly instead of failing on the unique index with a 500.
func lockClientID(tx *gorm.DB, table, clientID string) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	return tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", table+".client_id:"+clientID).Error
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// AddPlannedMealFromSavedMeal appends a planned copy of a saved meal. A non-empty clientID makes it
// idempotent: when a planned meal with that id exists, nothing is created.
func AddPlannedMealFromSavedMeal(ctx context.Context, offset int, savedMealID uint, meal *models.Meal, clientID string) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		day, err := findOrCreateDietDayForCalendarDateInTx(tx, clock.ZerodTime(offset))
		if err != nil {
			return err
		}
		if clientID != "" {
			if err := lockClientID(tx, "planned_meals", clientID); err != nil {
				return err
			}
			// Unscoped: the unique index covers soft-deleted planned meals too, so a replay
			// after the planned meal was removed is still a replay.
			var existing int64
			if err := tx.Unscoped().Model(&models.PlannedMeal{}).Where("client_id = ?", clientID).Count(&existing).Error; err != nil {
				return err
			}
			if existing > 0 {
				return nil
			}
		}
		for i := range meal.Items {
			meal.Items[i].ID = 0
		}
		if err := tx.Create(meal).Error; err != nil {
			return err
		}
		var next int
		if err := tx.Model(&models.PlannedMeal{}).
			Where("day_id = ? AND logged = ?", day.ID, false).
			Select("COALESCE(MAX(display_order), -1) + 1").
			Scan(&next).Error; err != nil {
			return err
		}
		sid := savedMealID
		pm := models.PlannedMeal{
			DayID:        day.ID,
			MealID:       meal.ID,
			SavedMealID:  &sid,
			Logged:       false,
			DisplayOrder: next,
			ClientID:     nilIfEmpty(clientID),
		}
		return tx.Create(&pm).Error
	})
}

func findOrCreateDietDayForCalendarDateInTx(tx *gorm.DB, t time.Time) (models.DietDay, error) {
	start, end := calendarDayRange(t)
	var day models.DietDay
	err := tx.Where("date >= ? AND date < ?", start, end).First(&day).Error
	if err == nil {
		return day, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.DietDay{}, err
	}
	plan, err := planForDate(tx, t)
	if err != nil {
		return models.DietDay{}, err
	}
	loc := t.Location()
	atMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	day = models.DietDay{
		Date:   atMidnight,
		PlanID: plan.ID,
	}
	if err := database.CreateInSavepoint(tx, &day); err != nil {
		if isUniqueConstraintError(err) {
			if err := tx.Where("date >= ? AND date < ?", start, end).First(&day).Error; err != nil {
				return models.DietDay{}, err
			}
			return day, nil
		}
		return models.DietDay{}, err
	}
	return day, nil
}

type QuickLogParams struct {
	DisplayName   string
	FoodRowName   string
	Calories      float32
	Protein       float32
	Carbs         float32
	Fat           float32
	Fiber         float32
	Offset        int
	ReplaceMealID uint
	// ClientID, when set, makes the call idempotent: if a meal with this id exists, nothing is created.
	ClientID string
}

func QuickLogMeal(ctx context.Context, params QuickLogParams) (dayID uint, err error) {
	err = database.From(ctx).Transaction(func(tx *gorm.DB) error {
		day, derr := findOrCreateDietDayForCalendarDateInTx(tx, clock.ZerodTime(params.Offset))
		if derr != nil {
			return derr
		}
		dayID = day.ID
		if params.ClientID != "" {
			if err := lockClientID(tx, "meals", params.ClientID); err != nil {
				return err
			}
			var existing int64
			if cerr := tx.Unscoped().Model(&models.Meal{}).Where("client_id = ?", params.ClientID).Count(&existing).Error; cerr != nil {
				return cerr
			}
			if existing > 0 {
				return nil
			}
		}
		food := models.Food{
			Name:          params.FoodRowName,
			ServingType:   "",
			ServingAmount: 1,
			Calories:      params.Calories,
			Protein:       params.Protein,
			Fiber:         params.Fiber,
			Carbs:         params.Carbs,
			Fat:           params.Fat,
			QuickEntry:    true,
		}
		if err := tx.Create(&food).Error; err != nil {
			return err
		}
		meal := models.Meal{
			Name:     params.DisplayName,
			ClientID: nilIfEmpty(params.ClientID),
			Items: []models.MealItem{{
				FoodID: food.ID,
				Amount: 1,
			}},
		}
		if err := tx.Create(&meal).Error; err != nil {
			return err
		}
		if params.ReplaceMealID != 0 {
			var cnt int64
			if cerr := tx.Model(&models.DayLog{}).
				Where("day_id = ? AND meal_id = ? AND deleted_at IS NULL", dayID, params.ReplaceMealID).
				Count(&cnt).Error; cerr != nil {
				return cerr
			}
			if cnt != 1 {
				return fmt.Errorf("replace_meal_id does not match a log on this day")
			}
			var log models.DayLog
			if err := tx.Where("day_id = ? AND meal_id = ?", dayID, params.ReplaceMealID).First(&log).Error; err != nil {
				return err
			}
			log.MealID = meal.ID
			return tx.Save(&log).Error
		}
		return tx.Create(&models.DayLog{
			DayID:  dayID,
			MealID: meal.ID,
		}).Error
	})
	return dayID, err
}

func EditLoggedMeal(ctx context.Context, dayID uint, oldMealID uint, meal *models.Meal) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		meal.ID = 0
		for i := range meal.Items {
			meal.Items[i].ID = 0
		}
		if err := tx.Create(meal).Error; err != nil {
			return err
		}
		var log models.DayLog
		if err := tx.Where("day_id = ? AND meal_id = ?", dayID, oldMealID).First(&log).Error; err != nil {
			return err
		}
		log.MealID = meal.ID
		return tx.Save(&log).Error
	})
}

// CopyPlannedMeals appends a copy of every unlogged planned meal on the source day (in its order) to
// each target day, after whatever that day already has. Each copy gets its own Meal row, so editing
// or logging one day never touches another. Targets equal to the source are skipped. Returns how many
// meals the source day had.
func CopyPlannedMeals(ctx context.Context, fromOffset int, toOffsets []int) (int, error) {
	var n int
	err := database.From(ctx).Transaction(func(tx *gorm.DB) error {
		src, err := findOrCreateDietDayForCalendarDateInTx(tx, clock.ZerodTime(fromOffset))
		if err != nil {
			return err
		}
		var planned []models.PlannedMeal
		if err := tx.Where("day_id = ? AND logged = ?", src.ID, false).
			Order("display_order ASC, id ASC").
			Preload("Meal.Items").
			Find(&planned).Error; err != nil {
			return err
		}
		n = len(planned)
		if n == 0 {
			return nil
		}
		seen := map[int]bool{fromOffset: true}
		for _, off := range toOffsets {
			if seen[off] {
				continue
			}
			seen[off] = true
			day, err := findOrCreateDietDayForCalendarDateInTx(tx, clock.ZerodTime(off))
			if err != nil {
				return err
			}
			var next int
			if err := tx.Model(&models.PlannedMeal{}).
				Where("day_id = ? AND logged = ?", day.ID, false).
				Select("COALESCE(MAX(display_order), -1) + 1").
				Scan(&next).Error; err != nil {
				return err
			}
			for i, p := range planned {
				meal := models.Meal{Name: p.Meal.Name}
				for _, it := range p.Meal.Items {
					meal.Items = append(meal.Items, models.MealItem{
						FoodID:          it.FoodID,
						Amount:          it.Amount,
						GroupID:         it.GroupID,
						GroupLabel:      it.GroupLabel,
						CompositeFoodID: it.CompositeFoodID,
					})
				}
				if err := tx.Create(&meal).Error; err != nil {
					return err
				}
				pm := models.PlannedMeal{
					DayID:        day.ID,
					MealID:       meal.ID,
					SavedMealID:  p.SavedMealID,
					DisplayOrder: next + i,
				}
				if err := tx.Create(&pm).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	return n, err
}
