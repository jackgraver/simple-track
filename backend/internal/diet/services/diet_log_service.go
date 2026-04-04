package services

import (
	"context"
	"time"

	"be-simpletracker/internal/diet/models"
	dietrepo "be-simpletracker/internal/diet/repository"

	"gorm.io/gorm"
)

// DietLogService coordinates diet log reads and derived totals.
type DietLogService struct {
	repo *dietrepo.Repository
}

func NewDietLogService(db *gorm.DB) *DietLogService {
	return &DietLogService{repo: dietrepo.New(db)}
}

func (s *DietLogService) MealPlanToday(_ context.Context, offset int) (models.Day, float32, float32, float32, float32, error) {
	day, err := s.repo.DayMealPlanToday(offset)
	if err != nil {
		return models.Day{}, 0, 0, 0, 0, err
	}
	tc, tp, tf, tb := s.repo.CalculateTotals(day.ID)
	return day, tc, tp, tf, tb, nil
}

func (s *DietLogService) MealPlanWeek(ctx context.Context) ([]models.Day, error) {
	today := time.Now()
	start := today.AddDate(0, 0, -3)
	end := today.AddDate(0, 0, 3)
	return s.repo.DaysByDateRange(ctx, start, end)
}

func (s *DietLogService) MealPlanMonth(ctx context.Context, offset int) (days []models.Day, startOfMonth, endOfMonth time.Time, month time.Month, err error) {
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

func (s *DietLogService) MealPlanDay(ctx context.Context, id uint) (models.Day, float32, float32, float32, float32, error) {
	day, err := s.repo.DayByIDGeneric(ctx, id)
	if err != nil {
		return models.Day{}, 0, 0, 0, 0, err
	}
	tc, tp, tf, tb := s.repo.CalculateTotals(day.ID)
	return day, tc, tp, tf, tb, nil
}

func (s *DietLogService) GoalsToday() (*models.Plan, error) {
	return s.repo.GoalsToday()
}
