package workoutrepo

import (
	"context"
	"time"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
)

func DatesWithLoggedSets(ctx context.Context, start, end time.Time) ([]time.Time, error) {
	var rows []struct {
		D time.Time `gorm:"column:d"`
	}
	err := conn().WithContext(ctx).Raw(`
		SELECT workout_logs.date AS d
		FROM logged_sets
		JOIN logged_exercises ON logged_exercises.id = logged_sets.logged_exercise_id
		JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id
		WHERE workout_logs.date >= ? AND workout_logs.date <= ?
		GROUP BY workout_logs.date
		ORDER BY workout_logs.date ASC
		`, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.D)
	}
	return out, nil
}

func DeleteLoggedSet(setID uint) error {
	return conn().Transaction(func(tx *gorm.DB) error {
		var set models.LoggedSet
		if err := tx.Where("id = ?", setID).First(&set).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Delete(&set).Error; err != nil {
			return err
		}

		var remainingSetCount int64
		if err := tx.Model(&models.LoggedSet{}).
			Where("logged_exercise_id = ?", set.LoggedExerciseID).
			Count(&remainingSetCount).Error; err != nil {
			return err
		}

		if remainingSetCount == 0 {
			if err := tx.Unscoped().
				Delete(&models.LoggedExercise{}, set.LoggedExerciseID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
