package repository

import (
	"context"
	"time"

	dbrepo "be-simpletracker/internal/database/repository"
	"be-simpletracker/internal/diet/models"
	"be-simpletracker/internal/utils"

	"gorm.io/gorm"
)

// Repository holds diet feature persistence (GORM).
type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FoodsAll(excludeIDs []uint) ([]models.Food, error) {
	var foods []models.Food
	query := r.db.Model(&models.Food{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Find(&foods).Error; err != nil {
		return nil, err
	}
	return foods, nil
}

func (r *Repository) FoodCreate(food *models.Food) error {
	return r.db.Create(food).Error
}

func (r *Repository) MealsAll(excludeIDs []uint) ([]models.Meal, error) {
	var meals []models.Meal
	query := r.db.Model(&models.Meal{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Preload("Items.Food").Find(&meals).Distinct("name").Error; err != nil {
		return nil, err
	}
	return meals, nil
}

func (r *Repository) MealByID(id uint) (*models.Meal, error) {
	var meal models.Meal
	if err := r.db.Preload("Items.Food").First(&meal, id).Error; err != nil {
		return nil, err
	}
	return &meal, nil
}

func (r *Repository) MealCreate(meal *models.Meal) (uint, error) {
	for i := range meal.Items {
		meal.Items[i].ID = 0
	}
	if err := r.db.Create(meal).Error; err != nil {
		return 0, err
	}
	return meal.ID, nil
}

func (r *Repository) DayMealPlanToday(offset int) (models.Day, error) {
	today := utils.ZerodTime(offset)
	var day models.Day
	if err := r.db.
		Preload("PlannedMeals", "logged = ?", false).
		Preload("PlannedMeals.Meal.Items.Food").
		Preload("Plan").
		Preload("Logs.Meal.Items.Food").
		Where("date = ?", today).
		First(&day).Error; err != nil {
		return models.Day{}, err
	}
	return day, nil
}

func (r *Repository) DayByID(id int) (*models.Day, error) {
	var day models.Day
	if err := r.db.
		Preload("PlannedMeals", "logged = ?", false).
		Preload("PlannedMeals.Meal.Items.Food").
		Preload("Plan").
		Preload("Logs.Meal.Items.Food").
		First(&day, id).Error; err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *Repository) DaysByDateRange(ctx context.Context, start, end time.Time) ([]models.Day, error) {
	repo := dbrepo.NewGormRepository[models.Day](r.db)
	return repo.GetByDateRange(ctx, start, end, dbrepo.WithDefaultPreloads())
}

func (r *Repository) DayByIDGeneric(ctx context.Context, id uint) (models.Day, error) {
	repo := dbrepo.NewGormRepository[models.Day](r.db)
	return repo.GetByID(ctx, id, dbrepo.WithDefaultPreloads())
}

func (r *Repository) CalculateTotals(dayID uint) (float32, float32, float32, float32) {
	var totals struct {
		TotalCalories float32 `json:"total_calories"`
		TotalProtein  float32 `json:"total_protein"`
		TotalFiber    float32 `json:"total_fiber"`
		TotalCarbs    float32 `json:"total_carbs"`
	}
	r.db.Raw(`
        SELECT
            SUM(f.calories * mi.amount) AS total_calories,
            SUM(f.protein  * mi.amount) AS total_protein,
            SUM(f.fiber    * mi.amount) AS total_fiber,
            SUM(f.carbs    * mi.amount) AS total_carbs
        FROM day_logs dl
        JOIN meals m       ON dl.meal_id = m.id
        JOIN meal_items mi ON mi.meal_id = m.id
        JOIN foods f       ON f.id = mi.food_id
        WHERE dl.day_id = ?
        AND dl.deleted_at IS NULL
    `, dayID).Scan(&totals)
	return totals.TotalCalories, totals.TotalProtein, totals.TotalFiber, totals.TotalCarbs
}

func (r *Repository) AllMealDays() ([]models.Day, error) {
	var days []models.Day
	if err := r.db.Find(&days).Error; err != nil {
		return nil, err
	}
	return days, nil
}

func (r *Repository) GoalsToday() (*models.Plan, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var todayDay models.Day
	if err := r.db.Where("date = ?", today.Format("2006-01-02")).First(&todayDay).Error; err != nil {
		return nil, err
	}
	var plan models.Plan
	if err := r.db.First(&plan, todayDay.PlanID).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *Repository) FindDayByDate(date time.Time) (*models.Day, error) {
	var day models.Day
	err := r.db.Where("date = ?", date).First(&day).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &day, nil
}

func (r *Repository) CreateDayMeal(dayMeal *models.DayLog) error {
	return r.db.Create(dayMeal).Error
}

func (r *Repository) SetPlannedMealLogged(dayID uint, mealID uint) error {
	var pm models.PlannedMeal
	if err := r.db.Where("day_id = ? AND meal_id = ?", dayID, mealID).First(&pm).Error; err != nil {
		return err
	}
	pm.Logged = true
	return r.db.Save(&pm).Error
}

func (r *Repository) DeleteLoggedMeal(dayID uint, mealID uint) error {
	var log models.DayLog
	if err := r.db.Where("day_id = ? AND meal_id = ?", dayID, mealID).First(&log).Error; err != nil {
		return err
	}
	return r.db.Delete(&log).Error
}

func (r *Repository) UpdateDayLogMeal(dayID uint, oldMealID uint, newMealID uint) error {
	var log models.DayLog
	if err := r.db.Where("day_id = ? AND meal_id = ?", dayID, oldMealID).First(&log).Error; err != nil {
		return err
	}
	log.MealID = newMealID
	return r.db.Save(&log).Error
}

func (r *Repository) PlansGetAll(ctx context.Context, params utils.QueryParams) (*utils.GetAllResult[models.Plan], error) {
	repo := dbrepo.NewGormRepository[models.Plan](r.db)
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
