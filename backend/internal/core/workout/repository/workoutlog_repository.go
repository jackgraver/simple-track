package workoutrepo

import (
	"context"
	"time"

	"be-simpletracker/internal/core/workout/models"
	dbrepo "be-simpletracker/internal/database/repository"

	"gorm.io/gorm"
)

func LoadByDate(ctx context.Context, date time.Time) (models.WorkoutLog, error) {
	var workoutDay models.WorkoutLog
	err := conn().WithContext(ctx).
		Preload("Cardio").
		Preload("Exercises.Sets").
		Preload("Exercises.Exercise").
		Preload("WorkoutPlan").
		Where("date = ?", date).
		First(&workoutDay).Error
	if err != nil {
		return models.WorkoutLog{}, err
	}
	if workoutDay.WorkoutPlan != nil {
		ex, err := LoadExercisesOrderedForPlan(workoutDay.WorkoutPlan.ID)
		if err != nil {
			return models.WorkoutLog{}, err
		}
		workoutDay.WorkoutPlan.Exercises = ex
	}
	return workoutDay, nil
}

func CreateMinimal(ctx context.Context, log *models.WorkoutLog) error {
	return conn().WithContext(ctx).Omit("WorkoutPlan", "Exercises", "Cardio").Create(log).Error
}

func GetByDateRange(ctx context.Context, start, end time.Time) ([]models.WorkoutLog, error) {
	repo := dbrepo.NewGormRepository[models.WorkoutLog](conn())
	return repo.GetByDateRange(ctx, start, end, dbrepo.WithDefaultPreloads())
}

func UpdateWorkoutPlanID(ctx context.Context, workoutLogID uint, planID *uint) error {
	var wid any
	if planID == nil {
		wid = nil
	} else {
		wid = *planID
	}
	return conn().WithContext(ctx).Model(&models.WorkoutLog{}).Where("id = ?", workoutLogID).Updates(map[string]any{
		"workout_plan_id": wid,
	}).Error
}

func UpdatePreMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := conn().WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PreMobilityChecked = checked
	return conn().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}

func UpdatePostMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := conn().WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PostMobilityChecked = checked
	return conn().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}
