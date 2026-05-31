package dietrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"be-simpletracker/internal/core/diet/models"
	dbrepo "be-simpletracker/internal/database/repository"
	"be-simpletracker/internal/utils"

	"gorm.io/gorm"
)

func EnrichFoodVariants(f *models.Food) {
	if f == nil {
		return
	}
	models.NormalizeQuickLogFoodNameForResponse(f)
	if f.VariantGroupID == nil {
		return
	}
	var sibs []models.Food
	if err := conn().Where("variant_group_id = ? AND id != ?", *f.VariantGroupID, f.ID).Order("name ASC").Find(&sibs).Error; err != nil {
		return
	}
	for i := range sibs {
		models.NormalizeQuickLogFoodNameForResponse(&sibs[i])
	}
	f.Variants = sibs
}

func enrichMealFoodVariants(m *models.Meal) {
	if m == nil {
		return
	}
	for i := range m.Items {
		if m.Items[i].Food.ID != 0 {
			EnrichFoodVariants(&m.Items[i].Food)
		}
	}
}

func enrichSavedMealFoodVariants(sm *models.SavedMeal) {
	if sm == nil {
		return
	}
	for i := range sm.Items {
		if sm.Items[i].Food.ID != 0 {
			EnrichFoodVariants(&sm.Items[i].Food)
		}
	}
}

func enrichDietDayFoodVariants(d *models.DietDay) {
	if d == nil {
		return
	}
	for i := range d.PlannedMeals {
		if d.PlannedMeals[i].Meal.ID != 0 {
			enrichMealFoodVariants(&d.PlannedMeals[i].Meal)
		}
	}
	for i := range d.Logs {
		if d.Logs[i].Meal.ID != 0 {
			enrichMealFoodVariants(&d.Logs[i].Meal)
		}
	}
}

func FoodsAll(excludeIDs []uint) ([]models.Food, error) {
	var foods []models.Food
	query := conn().Model(&models.Food{}).Where("quick_entry = ? OR quick_entry IS NULL", false).Order("name ASC")
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Find(&foods).Error; err != nil {
		return nil, err
	}
	return foods, nil
}

func FoodCreate(food *models.Food) error {
	return conn().Create(food).Error
}

// FoodCreateWithOptionalRelated inserts a food and, if set, links it to a related food's variant group.
func FoodCreateWithOptionalRelated(food *models.Food, relatedFoodID *uint) error {
	return conn().Transaction(func(tx *gorm.DB) error {
		food.ID = 0
		if err := tx.Create(food).Error; err != nil {
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

func foodsByVariantGroup() (map[uint][]models.Food, error) {
	var list []models.Food
	if err := conn().Model(&models.Food{}).Where("variant_group_id IS NOT NULL").Order("name ASC").Find(&list).Error; err != nil {
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
func FoodsAllWithVariantSiblings(excludeIDs []uint) ([]models.FoodWithVariants, error) {
	foods, err := FoodsAll(excludeIDs)
	if err != nil {
		return nil, err
	}
	byGroup, err := foodsByVariantGroup()
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

func CompositeFoodsAll() ([]models.CompositeFood, error) {
	var list []models.CompositeFood
	if err := conn().Preload("Items.Food").Order("name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		for j := range list[i].Items {
			models.NormalizeQuickLogFoodNameForResponse(&list[i].Items[j].Food)
		}
	}
	return list, nil
}

func CompositeFoodCreate(cf *models.CompositeFood) (uint, error) {
	cf.ID = 0
	for i := range cf.Items {
		cf.Items[i].ID = 0
	}
	if err := conn().Create(cf).Error; err != nil {
		return 0, err
	}
	return cf.ID, nil
}

func MealsAll(excludeIDs []uint) ([]models.Meal, error) {
	var meals []models.Meal
	query := conn().Model(&models.Meal{}).Order("name ASC")
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Preload("Items.Food").Find(&meals).Distinct("name").Error; err != nil {
		return nil, err
	}
	return meals, nil
}

func SavedMealsAll(excludeIDs []uint) ([]models.SavedMeal, error) {
	var meals []models.SavedMeal
	query := conn().Model(&models.SavedMeal{}).Order("name ASC")
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Preload("Items.Food").Find(&meals).Distinct("name").Error; err != nil {
		return nil, err
	}
	for i := range meals {
		if meals[i].Items != nil {
			enrichSavedMealFoodVariants(&meals[i])
		}
	}
	return meals, nil
}

func SavedMealCreate(sm *models.SavedMeal) (uint, error) {
	for i := range sm.Items {
		sm.Items[i].ID = 0
	}
	if err := conn().Create(sm).Error; err != nil {
		return 0, err
	}
	return sm.ID, nil
}

func SavedMealByID(id uint) (*models.SavedMeal, error) {
	var sm models.SavedMeal
	if err := conn().Preload("Items.Food").First(&sm, id).Error; err != nil {
		return nil, err
	}
	enrichSavedMealFoodVariants(&sm)
	return &sm, nil
}

func SavedMealDelete(id uint) error {
	var sm models.SavedMeal
	if err := conn().First(&sm, id).Error; err != nil {
		return err
	}
	return conn().Delete(&sm).Error
}

func CountUnloggedPlannedBySavedMealID(savedMealID uint) (int64, error) {
	var count int64
	err := conn().Model(&models.PlannedMeal{}).
		Where("saved_meal_id = ? AND logged = ?", savedMealID, false).
		Count(&count).Error
	return count, err
}

func DeleteUnloggedPlannedBySavedMealID(savedMealID uint) error {
	return conn().Where("saved_meal_id = ? AND logged = ?", savedMealID, false).Delete(&models.PlannedMeal{}).Error
}

// SavedMealReplace updates the template name and replaces all items in one transaction.
func SavedMealReplace(id uint, incoming *models.SavedMeal) error {
	return conn().Transaction(func(tx *gorm.DB) error {
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

func PlannedMealCreate(pm *models.PlannedMeal) error {
	pm.ID = 0
	return conn().Create(pm).Error
}

func PlannedMealDelete(plannedMealID uint, dayID uint) error {
	var pm models.PlannedMeal
	if err := conn().Where("id = ? AND day_id = ?", plannedMealID, dayID).First(&pm).Error; err != nil {
		return err
	}
	return conn().Delete(&pm).Error
}

func NextPlannedMealDisplayOrder(dayID uint) (int, error) {
	var next int
	err := conn().Model(&models.PlannedMeal{}).
		Where("day_id = ? AND logged = ?", dayID, false).
		Select("COALESCE(MAX(display_order), -1) + 1").
		Scan(&next).Error
	return next, err
}

func PlannedMealReorder(dayID uint, orderedIDs []uint) error {
	return conn().Transaction(func(tx *gorm.DB) error {
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

func MealByID(id uint) (*models.Meal, error) {
	var meal models.Meal
	if err := conn().Preload("Items.Food").First(&meal, id).Error; err != nil {
		return nil, err
	}
	enrichMealFoodVariants(&meal)
	return &meal, nil
}

func MealCreate(meal *models.Meal) (uint, error) {
	for i := range meal.Items {
		meal.Items[i].ID = 0
	}
	if err := conn().Create(meal).Error; err != nil {
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
func findOrCreateDietDayForCalendarDate(t time.Time) (models.DietDay, error) {
	start, end := calendarDayRange(t)
	var day models.DietDay
	err := conn().Where("date >= ? AND date < ?", start, end).First(&day).Error
	if err == nil {
		return day, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.DietDay{}, err
	}

	plan, err := planForDate(conn(), t)
	if err != nil {
		return models.DietDay{}, err
	}

	loc := t.Location()
	atMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	day = models.DietDay{
		Date:   atMidnight,
		PlanID: plan.ID,
	}
	if err := conn().Create(&day).Error; err != nil {
		if isUniqueConstraintError(err) {
			if err := conn().Where("date >= ? AND date < ?", start, end).First(&day).Error; err != nil {
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

func loadDietDayWithPreloads(id uint) (models.DietDay, error) {
	var day models.DietDay
	if err := conn().
		Preload("PlannedMeals", func(db *gorm.DB) *gorm.DB {
			return db.Where("logged = ?", false).Order("display_order ASC, id ASC")
		}).
		Preload("PlannedMeals.Meal.Items.Food").
		Preload("Plan").
		Preload("Logs.Meal.Items.Food").
		First(&day, id).Error; err != nil {
		return models.DietDay{}, err
	}
	enrichDietDayFoodVariants(&day)
	return day, nil
}

func DayMealPlanToday(offset int) (models.DietDay, error) {
	d, err := findOrCreateDietDayForCalendarDate(utils.ZerodTime(offset))
	if err != nil {
		return models.DietDay{}, err
	}
	return loadDietDayWithPreloads(d.ID)
}

// CountUnloggedPlannedMealsPerCalendarDay returns YYYY-MM-DD (loc) → count for that calendar day.
func CountUnloggedPlannedMealsPerCalendarDay(start, end time.Time) (map[string]int, error) {
	out := make(map[string]int)
	rows, err := conn().Model(&models.PlannedMeal{}).
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

func DayByID(id int) (*models.DietDay, error) {
	day, err := loadDietDayWithPreloads(uint(id))
	if err != nil {
		return nil, err
	}
	return &day, nil
}

func DaysByDateRange(ctx context.Context, start, end time.Time) ([]models.DietDay, error) {
	repo := dbrepo.NewGormRepository[models.DietDay](conn())
	return repo.GetByDateRange(ctx, start, end, dbrepo.WithDefaultPreloads())
}

func DayByIDGeneric(ctx context.Context, id uint) (models.DietDay, error) {
	repo := dbrepo.NewGormRepository[models.DietDay](conn())
	return repo.GetByID(ctx, id, dbrepo.WithDefaultPreloads())
}

// MealDayTotals aggregates logged macros for one diet day from day_logs.
type MealDayTotals struct {
	Calories float32
	Protein  float32
	Fiber    float32
	Carbs    float32
	Fat      float32
}

func CalculateTotals(dayID uint) MealDayTotals {
	var totals MealDayTotals
	conn().Raw(`
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
    `, dayID).Scan(&totals)
	return totals
}

func AllMealDays() ([]models.DietDay, error) {
	var days []models.DietDay
	if err := conn().Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

func GoalsToday() (*models.Plan, error) {
	todayDay, err := findOrCreateDietDayForCalendarDate(utils.ZerodTime(0))
	if err != nil {
		return nil, err
	}
	var plan models.Plan
	if err := conn().First(&plan, todayDay.PlanID).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func FindDayByDate(date time.Time) (*models.DietDay, error) {
	day, err := findOrCreateDietDayForCalendarDate(date)
	if err != nil {
		return nil, err
	}
	return &day, nil
}

func CreateDayMeal(dayMeal *models.DayLog) error {
	return conn().Create(dayMeal).Error
}

func DayLogExists(dayID uint, mealID uint) (bool, error) {
	var count int64
	err := conn().Model(&models.DayLog{}).
		Where("day_id = ? AND meal_id = ?", dayID, mealID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func SetPlannedMealLogged(dayID uint, mealID uint) error {
	var pm models.PlannedMeal
	if err := conn().Where("day_id = ? AND meal_id = ?", dayID, mealID).First(&pm).Error; err != nil {
		return err
	}
	pm.Logged = true
	return conn().Save(&pm).Error
}

func DeleteLoggedMeal(dayID uint, mealID uint) error {
	var log models.DayLog
	if err := conn().Where("day_id = ? AND meal_id = ?", dayID, mealID).First(&log).Error; err != nil {
		return err
	}
	return conn().Delete(&log).Error
}

func UpdateDayLogMeal(dayID uint, oldMealID uint, newMealID uint) error {
	var log models.DayLog
	if err := conn().Where("day_id = ? AND meal_id = ?", dayID, oldMealID).First(&log).Error; err != nil {
		return err
	}
	log.MealID = newMealID
	return conn().Save(&log).Error
}

func PlansGetAll(ctx context.Context, params utils.QueryParams) (*utils.GetAllResult[models.Plan], error) {
	repo := dbrepo.NewGormRepository[models.Plan](conn())
	opts := utils.BuildQueryOptions(params, "id", true)
	if params.Page > 0 && params.PageSize > 0 {
		result, err := repo.GetAllPaginated(ctx, params.Page, params.PageSize, opts...)
		if err != nil {
			return nil, err
		}
		return &utils.GetAllResult[models.Plan]{
			Data:       result.Data,
			Pagination: result,
		}, nil
	}
	entities, err := repo.GetAll(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &utils.GetAllResult[models.Plan]{
		Data:       entities,
		Pagination: nil,
	}, nil
}

func CompositeFoodByID(id uint) (*models.CompositeFood, error) {
	var cf models.CompositeFood
	if err := conn().Preload("Items.Food").First(&cf, id).Error; err != nil {
		return nil, err
	}
	for i := range cf.Items {
		models.NormalizeQuickLogFoodNameForResponse(&cf.Items[i].Food)
	}
	return &cf, nil
}

func UpdatePlanMacros(id uint, calories, protein, fiber, carbs, fat float32) (*models.Plan, error) {
	effectiveStart, _ := calendarDayRange(utils.ZerodTime(0))
	var plan models.Plan

	if err := conn().Transaction(func(tx *gorm.DB) error {
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

func DeleteSavedMeal(id uint) error {
	return conn().Transaction(func(tx *gorm.DB) error {
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

func addPlannedMealInTx(tx *gorm.DB, offset int, meal *models.Meal, savedMealID *uint) error {
	day, err := findOrCreateDietDayForCalendarDateInTx(tx, utils.ZerodTime(offset))
	if err != nil {
		return err
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
	pm := models.PlannedMeal{
		DayID:        day.ID,
		MealID:       meal.ID,
		SavedMealID:  savedMealID,
		Logged:       false,
		DisplayOrder: next,
	}
	return tx.Create(&pm).Error
}

func AddPlannedMealFromSavedMeal(offset int, savedMealID uint, meal *models.Meal) error {
	sid := savedMealID
	return conn().Transaction(func(tx *gorm.DB) error {
		return addPlannedMealInTx(tx, offset, meal, &sid)
	})
}

func AddPlannedMealFromLabel(offset int, meal *models.Meal) error {
	return conn().Transaction(func(tx *gorm.DB) error {
		return addPlannedMealInTx(tx, offset, meal, nil)
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
	if err := tx.Create(&day).Error; err != nil {
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
}

func QuickLogMeal(params QuickLogParams) (dayID uint, err error) {
	err = conn().Transaction(func(tx *gorm.DB) error {
		day, derr := findOrCreateDietDayForCalendarDateInTx(tx, utils.ZerodTime(params.Offset))
		if derr != nil {
			return derr
		}
		dayID = day.ID
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
			Name: params.DisplayName,
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

func EditLoggedMeal(dayID uint, oldMealID uint, meal *models.Meal) error {
	return conn().Transaction(func(tx *gorm.DB) error {
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
