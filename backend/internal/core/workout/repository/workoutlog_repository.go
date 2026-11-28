package workoutrepo

import (
	"be-simpletracker/internal/database"
	"context"
	"time"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func LoadByDate(ctx context.Context, date time.Time) (models.WorkoutLog, error) {
	own := database.Owned(ctx)
	var workoutDay models.WorkoutLog
	err := database.From(ctx).WithContext(ctx).Scopes(own).
		Preload("Cardio", own).
		Preload("Exercises", own).
		Preload("Exercises.Sets", func(db *gorm.DB) *gorm.DB { return setsInLoggedOrder(own(db)) }).
		Preload("Exercises.Exercise").
		// Unscoped: a routine deleted after the session was logged still names and shapes that log.
		Preload("WorkoutPlan", func(db *gorm.DB) *gorm.DB { return own(db.Unscoped()) }).
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
	var logged []*models.Exercise
	for i := range workoutDay.Exercises {
		if workoutDay.Exercises[i].Exercise != nil {
			logged = append(logged, workoutDay.Exercises[i].Exercise)
		}
	}
	if err := DecorateExercises(ctx, logged...); err != nil {
		return models.WorkoutLog{}, err
	}
	return workoutDay, nil
}

// CreateMinimalIfMissing inserts log unless the caller already has a live workout log for
// log.Date. The check-and-insert is serialized per user and date with a transaction-scoped
// advisory lock (database.XactLock puts the user in the key, so two users never wait on each
// other), and the count only sees the caller's rows. If a row already exists, log is left without
// an ID and nil is returned. idx_workout_logs_user_date (migration 0017) is the backstop.
func CreateMinimalIfMissing(ctx context.Context, log *models.WorkoutLog) error {
	return database.From(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := database.XactLock(ctx, tx, "workout_logs.date", log.Date.Format("2006-01-02")); err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&models.WorkoutLog{}).Scopes(database.Owned(ctx)).Where("date = ?", log.Date).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		return tx.Omit(clause.Associations).Create(log).Error
	})
}

func GetByDateRange(ctx context.Context, start, end time.Time) ([]models.WorkoutLog, error) {
	own := database.Owned(ctx)
	var logs []models.WorkoutLog
	err := database.From(ctx).WithContext(ctx).Scopes(own).
		Preload("Cardio", own).
		Preload("Exercises", own).
		Preload("Exercises.Sets", own).
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
	return database.Affected(database.From(ctx).WithContext(ctx).Model(&models.WorkoutLog{}).Scopes(database.Owned(ctx)).
		Where("id = ?", workoutLogID).Updates(map[string]any{
		"workout_plan_id": wid,
	}))
}

func UpdatePreMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	return updateMobilityChecked(ctx, workoutLogID, "pre_mobility_checked", checked)
}

func UpdatePostMobilityChecked(ctx context.Context, workoutLogID uint, checked []string) error {
	return updateMobilityChecked(ctx, workoutLogID, "post_mobility_checked", checked)
}

func updateMobilityChecked(ctx context.Context, workoutLogID uint, column string, checked []string) error {
	var wl models.WorkoutLog
	if err := database.From(ctx).WithContext(ctx).Scopes(database.Owned(ctx)).First(&wl, workoutLogID).Error; err != nil {
		return err
	}
	if column == "pre_mobility_checked" {
		wl.PreMobilityChecked = checked
	} else {
		wl.PostMobilityChecked = checked
	}
	// Only the one column: Save would rewrite every column, user_id included.
	field := map[string]string{"pre_mobility_checked": "PreMobilityChecked", "post_mobility_checked": "PostMobilityChecked"}[column]
	return database.Affected(database.From(ctx).WithContext(ctx).Model(&wl).Scopes(database.Owned(ctx)).
		Select(field).Updates(&wl))
}
