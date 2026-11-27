package workoutrepo

import (
	"be-simpletracker/internal/database"
	"context"
	"time"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
)

func LoadByDate(ctx context.Context, date time.Time) (models.WorkoutLog, error) {
	var workoutDay models.WorkoutLog
	err := database.From(ctx).WithContext(ctx).
		Preload("Cardio").
		Preload("Exercises.Sets", setsInLoggedOrder).
		Preload("Exercises.Exercise").
		// Unscoped: a routine deleted after the session was logged still names and shapes that log.
		Preload("WorkoutPlan", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Where("date = ?", date).
		First(&workoutDay).Error
	if err != nil {
		return models.WorkoutLog{}, err
	}
	if workoutDay.WorkoutPlan != nil {
		ex, err := LoadExercisesOrderedForPlan(ctx, workoutDay.WorkoutPlan.ID)
		if err != nil {
			return models.WorkoutLog{}, err
		}
		workoutDay.WorkoutPlan.Exercises = ex
	}
	return workoutDay, nil
}

// CreateMinimalIfMissing inserts log unless a live workout log already exists for log.Date.
// workout_logs.date has no unique index (legacy duplicates may exist), so on Postgres the
// check-and-insert is serialized per date with a transaction-scoped advisory lock. If a row
// already exists, log is left without an ID and nil is returned.
func CreateMinimalIfMissing(ctx context.Context, log *models.WorkoutLog) error {
	return database.From(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			key := "workout_logs.date:" + log.Date.Format("2006-01-02")
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", key).Error; err != nil {
				return err
			}
		}
		var count int64
		if err := tx.Model(&models.WorkoutLog{}).Where("date = ?", log.Date).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		return tx.Omit("WorkoutPlan", "Exercises", "Cardio").Create(log).Error
	})
}

func GetByDateRange(ctx context.Context, start, end time.Time) ([]models.WorkoutLog, error) {
	var logs []models.WorkoutLog
	err := database.From(ctx).WithContext(ctx).
		Preload("Cardio").
		Preload("Exercises.Sets").
		Preload("Exercises.Exercise").
		Where("date >= ? AND date <= ?", start, end).
		Order("date ASC").
		Find(&logs).Error
	return logs, err
}

func UpdateWorkoutPlanID(ctx context.Context, workoutLogID uint, planID *uint) error {
	var wid any
	if planID == nil {
		wid = nil
	} else {
		wid = *planID
	}
	return database.From(ctx).WithContext(ctx).Model(&models.WorkoutLog{}).Where("id = ?", workoutLogID).Updates(map[string]any{
		"workout_plan_id": wid,
	}).Error
}

func UpdatePreMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := database.From(ctx).WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PreMobilityChecked = checked
	return database.From(ctx).WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}

func UpdatePostMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	var wl models.WorkoutLog
	if err := database.From(ctx).WithContext(ctx).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	wl.PostMobilityChecked = checked
	return database.From(ctx).WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: false}).Save(&wl).Error
}
