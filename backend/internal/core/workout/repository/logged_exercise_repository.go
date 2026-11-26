package workoutrepo

import (
	"context"
	"errors"
	"time"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrLoggedExerciseExists is returned by CreateLoggedExercise when the workout log already has a
// live row for that exercise. The exercise's ID is set to the existing row.
var ErrLoggedExerciseExists = errors.New("exercise already logged in this workout")

// liveLoggedExerciseConflict targets idx_logged_exercises_workout_exercise (partial, live rows only).
func liveLoggedExerciseConflict() clause.OnConflict {
	return clause.OnConflict{
		Columns:     []clause.Column{{Name: "workout_log_id"}, {Name: "exercise_id"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "deleted_at IS NULL"}}},
		DoNothing:   true,
	}
}

// upsertLoggedExercise returns the live logged_exercise for (workoutLogID, exerciseID), inserting
// it if missing. Safe under concurrency: the unique index decides which insert wins and the loser
// reads the winner's row.
func upsertLoggedExercise(tx *gorm.DB, workoutLogID, exerciseID uint, notes string) (uint, bool, error) {
	le := models.LoggedExercise{WorkoutLogID: workoutLogID, ExerciseID: exerciseID, Notes: notes}
	res := tx.Omit("Exercise", "Sets").Clauses(liveLoggedExerciseConflict()).Create(&le)
	if res.Error != nil {
		return 0, false, res.Error
	}
	if res.RowsAffected == 1 && le.ID != 0 {
		return le.ID, true, nil
	}
	var existing models.LoggedExercise
	if err := tx.Select("id").
		Where("workout_log_id = ? AND exercise_id = ?", workoutLogID, exerciseID).
		First(&existing).Error; err != nil {
		return 0, false, err
	}
	return existing.ID, false, nil
}

// CreateLoggedExercise inserts a logged exercise and its sets. If the workout log already has a
// live row for the exercise, nothing is written and ErrLoggedExerciseExists is returned with
// exercise.ID set to the existing row.
func CreateLoggedExercise(exercise *models.LoggedExercise) error {
	return conn().Transaction(func(tx *gorm.DB) error {
		id, created, err := upsertLoggedExercise(tx, exercise.WorkoutLogID, exercise.ExerciseID, exercise.Notes)
		if err != nil {
			return err
		}
		exercise.ID = id
		if !created {
			return ErrLoggedExerciseExists
		}
		for i := range exercise.Sets {
			exercise.Sets[i].ID = 0
			exercise.Sets[i].LoggedExerciseID = id
			if err := tx.Create(&exercise.Sets[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func UpdateLoggedExerciseWithSets(exercise models.LoggedExercise) error {
	return conn().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.LoggedExercise{}).
			Where("id = ?", exercise.ID).
			Updates(map[string]any{
				"workout_log_id": exercise.WorkoutLogID,
				"exercise_id":    exercise.ExerciseID,
				"notes":          exercise.Notes,
			}).Error; err != nil {
			return err
		}

		incomingSetIDs := make(map[uint]struct{}, len(exercise.Sets))
		for i := range exercise.Sets {
			set := exercise.Sets[i]
			set.LoggedExerciseID = exercise.ID

			if set.ID > 0 {
				incomingSetIDs[set.ID] = struct{}{}
				if err := tx.Model(&models.LoggedSet{}).
					Where("id = ? AND logged_exercise_id = ?", set.ID, exercise.ID).
					Updates(map[string]any{
						"reps":               set.Reps,
						"weight":             set.Weight,
						"weight_setup":       set.WeightSetup,
						"logged_exercise_id": exercise.ID,
					}).Error; err != nil {
					return err
				}
				continue
			}

			set.ID = 0
			if err := tx.Create(&set).Error; err != nil {
				return err
			}
			incomingSetIDs[set.ID] = struct{}{}
		}

		var existingSetIDs []uint
		if err := tx.Model(&models.LoggedSet{}).
			Where("logged_exercise_id = ?", exercise.ID).
			Pluck("id", &existingSetIDs).Error; err != nil {
			return err
		}

		setIDsToDelete := make([]uint, 0)
		for _, existingSetID := range existingSetIDs {
			if _, ok := incomingSetIDs[existingSetID]; !ok {
				setIDsToDelete = append(setIDsToDelete, existingSetID)
			}
		}

		if len(setIDsToDelete) > 0 {
			if err := tx.Unscoped().
				Where("logged_exercise_id = ? AND id IN ?", exercise.ID, setIDsToDelete).
				Delete(&models.LoggedSet{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func RemoveLoggedExerciseForDay(ctx context.Context, day time.Time, exerciseID uint) error {
	res := conn().WithContext(ctx).Unscoped().
		Where(
			"exercise_id = ? AND workout_log_id IN (?)",
			exerciseID,
			conn().Model(&models.WorkoutLog{}).Select("id").Where("date = ?", day),
		).
		Delete(&models.LoggedExercise{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func LoadLoggedExercise(id uint) (models.LoggedExercise, error) {
	var exercise models.LoggedExercise
	err := conn().Preload("Exercise").
		Preload("Sets", setsInLoggedOrder).
		Where("id = ?", id).First(&exercise).Error
	if err != nil {
		return models.LoggedExercise{}, err
	}
	if err := attachWorkoutLogDate(&exercise); err != nil {
		return models.LoggedExercise{}, err
	}
	return exercise, nil
}

func attachWorkoutLogDate(exerciseLog *models.LoggedExercise) error {
	if exerciseLog == nil || exerciseLog.ID == 0 {
		return nil
	}
	var workoutLog models.WorkoutLog
	err := conn().Select("date").First(&workoutLog, exerciseLog.WorkoutLogID).Error
	if err != nil {
		return err
	}
	exerciseLog.LogDate = workoutLog.Date
	return nil
}

func GetPreviousExerciseLog(ctx context.Context, day time.Time, exercise string, offset int) (models.LoggedExercise, error) {
	var exerciseLog models.LoggedExercise
	err := conn().WithContext(ctx).
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id").
		Joins("JOIN exercises ON exercises.id = logged_exercises.exercise_id").
		Where("exercises.name = ?", exercise).
		Where("workout_logs.date != ?", day).
		Where("workout_logs.date < ?", day).
		Preload("Sets", setsInLoggedOrder).
		Preload("Exercise").
		Order("workout_logs.date DESC").
		Offset(offset).
		Limit(1).
		Find(&exerciseLog).Error
	if err != nil {
		return models.LoggedExercise{}, err
	}
	if exerciseLog.ID == 0 {
		return exerciseLog, nil
	}
	if err := attachWorkoutLogDate(&exerciseLog); err != nil {
		return models.LoggedExercise{}, err
	}
	return exerciseLog, nil
}

func GetMaxExerciseLog(ctx context.Context, day time.Time, exercise string) (models.LoggedExercise, error) {
	var exerciseLog models.LoggedExercise
	err := conn().WithContext(ctx).
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id").
		Joins("JOIN exercises ON exercises.id = logged_exercises.exercise_id").
		Joins("JOIN logged_sets ON logged_sets.logged_exercise_id = logged_exercises.id").
		Where("exercises.name = ?", exercise).
		Where("workout_logs.date != ?", day).
		Where("workout_logs.date < ?", day).
		Where("workout_logs.deleted_at IS NULL").
		Where("exercises.deleted_at IS NULL").
		Where("logged_sets.deleted_at IS NULL").
		Preload("Sets", setsInLoggedOrder).
		Preload("Exercise").
		Order("logged_sets.weight DESC").
		Order("logged_sets.reps DESC").
		Order("workout_logs.date DESC").
		Limit(1).
		First(&exerciseLog).Error
	if err != nil {
		return models.LoggedExercise{}, err
	}
	if exerciseLog.ID == 0 {
		return exerciseLog, nil
	}
	if err := attachWorkoutLogDate(&exerciseLog); err != nil {
		return models.LoggedExercise{}, err
	}
	return exerciseLog, nil
}

// setsInLoggedOrder preloads sets in the order they were logged. Progression reads the final set as
// where the next session starts, and Postgres returns unordered rows in heap order, which an edit can change.
func setsInLoggedOrder(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }
