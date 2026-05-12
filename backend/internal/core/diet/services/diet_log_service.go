package services

import (
	"context"
	"time"

	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"

	"gorm.io/gorm"
)

// DietLogService coordinates diet log reads and derived totals.
type DietLogService struct {
	repo *dietrepo.Repository
}

func NewDietLogService(db *gorm.DB) *DietLogService {
	return &DietLogService{repo: dietrepo.New(db)}
}

func (s *DietLogService) MealPlanToday(_ context.Context, offset int) (models.DietDay, dietrepo.MealDayTotals, error) {
	day, err := s.repo.DayMealPlanToday(offset)
	if err != nil {
		return models.DietDay{}, dietrepo.MealDayTotals{}, err
	}
	tot := s.repo.CalculateTotals(day.ID)
	return day, tot, nil
}

func (s *DietLogService) MealPlanWeek(ctx context.Context) ([]models.DietDay, error) {
	today := time.Now()
	start := today.AddDate(0, 0, -3)
	end := today.AddDate(0, 0, 3)
	return s.repo.DaysByDateRange(ctx, start, end)
}

func (s *DietLogService) MealPlanMonth(ctx context.Context, offset int) (days []models.DietDay, startOfMonth, endOfMonth time.Time, month time.Month, err error) {
	today := time.Now()
	target := today.AddDate(0, offset, 0)
	startOfMonth = time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
	endOfMonth = startOfMonth.AddDate(0, 1, -1)
	days, err = s.repo.DaysByDateRange(ctx, startOfMonth, endOfMonth)
	if err != nil {
		return nil, startOfMonth, endOfMonth, 0, err
	}
	return days, startOfMonth, endOfMonth, target.Month(), nil
}

// MonthPlannedSummary returns one count per calendar day in the month (index 0 = day 1), only unlogged planned meals.
func (s *DietLogService) MonthPlannedSummary(_ context.Context, monthOffset int) ([]int, error) {
	today := time.Now()
	target := today.AddDate(0, monthOffset, 0)
	loc := target.Location()
	startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, loc)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	dim := endOfMonth.Day()
	byDate, err := s.repo.CountUnloggedPlannedMealsPerCalendarDay(startOfMonth, endOfMonth)
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

func (s *DietLogService) MealPlanDay(ctx context.Context, id uint) (models.DietDay, dietrepo.MealDayTotals, error) {
	day, err := s.repo.DayByIDGeneric(ctx, id)
	if err != nil {
		return models.DietDay{}, dietrepo.MealDayTotals{}, err
	}
	tot := s.repo.CalculateTotals(day.ID)
	return day, tot, nil
}

func (s *DietLogService) GoalsToday() (*models.Plan, error) {
	return s.repo.GoalsToday()
}
