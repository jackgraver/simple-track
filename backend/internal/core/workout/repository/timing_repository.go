package workoutrepo

import (
	"context"
	"time"
)

// TimedSetRow is one logged set with the context needed for timing stats.
// SetCreatedAt is when the set was saved, which the logging UI does as soon
// as the set is finished.
type TimedSetRow struct {
	WorkoutLogID     uint      `gorm:"column:workout_log_id"`
	LogDate          time.Time `gorm:"column:log_date"`
	LoggedExerciseID uint      `gorm:"column:logged_exercise_id"`
	ExerciseID       uint      `gorm:"column:exercise_id"`
	ExerciseName     string    `gorm:"column:exercise_name"`
	SetCreatedAt     time.Time `gorm:"column:set_created_at"`
}

// TimedSetsSince returns every logged set from workouts dated on or after
// start, ordered by workout then set creation time.
func TimedSetsSince(ctx context.Context, start time.Time) ([]TimedSetRow, error) {
	var rows []TimedSetRow
	err := conn().WithContext(ctx).Raw(`
		SELECT
			workout_logs.id AS workout_log_id,
			workout_logs.date AS log_date,
			logged_exercises.id AS logged_exercise_id,
			exercises.id AS exercise_id,
			exercises.name AS exercise_name,
			logged_sets.created_at AS set_created_at
		FROM logged_sets
		JOIN logged_exercises ON logged_exercises.id = logged_sets.logged_exercise_id
		JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id
		JOIN exercises ON exercises.id = logged_exercises.exercise_id
		WHERE workout_logs.date >= ?
			AND logged_sets.created_at IS NOT NULL
			AND logged_sets.deleted_at IS NULL
			AND logged_exercises.deleted_at IS NULL
			AND workout_logs.deleted_at IS NULL
		ORDER BY workout_logs.id ASC, logged_sets.created_at ASC, logged_sets.id ASC
		`, start).Scan(&rows).Error
	return rows, err
}
