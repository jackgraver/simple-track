package workoutrepo

import (
	"be-simpletracker/internal/database"
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
func upsertLoggedExercise(ctx context.Context, tx *gorm.DB, workoutLogID, exerciseID uint, notes string) (uint, bool, error) {
	le := models.LoggedExercise{WorkoutLogID: workoutLogID, ExerciseID: exerciseID, Notes: notes}
	res := tx.Omit(clause.Associations).Clauses(liveLoggedExerciseConflict()).Create(&le)
	if res.Error != nil {
		return 0, false, res.Error
	}
	if res.RowsAffected == 1 && le.ID != 0 {
		return le.ID, true, nil
	}
	var existing models.LoggedExercise
	if err := tx.Select("id").Scopes(database.Owned(ctx)).
		Where("workout_log_id = ? AND exercise_id = ?", workoutLogID, exerciseID).
		First(&existing).Error; err != nil {
		return 0, false, err
	}
	return existing.ID, false, nil
}

// requireOwnLogAndExercise is the up-front check for requests that name a workout log and an
// exercise: the log must be the caller's (404 otherwise, the same as a missing one) and the
// exercise must exist in the catalog. The composite FK would refuse another user's log anyway,
// but with a database error instead of a 404.
func requireOwnLogAndExercise(ctx context.Context, tx *gorm.DB, workoutLogID, exerciseID uint) error {
	if err := tx.Model(&models.WorkoutLog{}).Scopes(database.Owned(ctx)).Select("id").First(&models.WorkoutLog{}, workoutLogID).Error; err != nil {
		return err
	}
	// Unscoped: an exercise deleted from the catalog still takes logs for old sessions.
	return tx.Unscoped().Model(&models.Exercise{}).Select("id").First(&models.Exercise{}, exerciseID).Error
}

// CreateLoggedExercise inserts a logged exercise and its sets. If the workout log already has a
// live row for the exercise, nothing is written and ErrLoggedExerciseExists is returned with
// exercise.ID set to the existing row.
func CreateLoggedExercise(ctx context.Context, exercise *models.LoggedExercise) error {
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireOwnLogAndExercise(ctx, tx, exercise.WorkoutLogID, exercise.ExerciseID); err != nil {
			return err
		}
		id, created, err := upsertLoggedExercise(ctx, tx, exercise.WorkoutLogID, exercise.ExerciseID, exercise.Notes)
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
			if err := tx.Omit(clause.Associations).Create(&exercise.Sets[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func UpdateLoggedExerciseWithSets(ctx context.Context, exercise models.LoggedExercise) error {
	own := database.Owned(ctx)
	return database.From(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireOwnLogAndExercise(ctx, tx, exercise.WorkoutLogID, exercise.ExerciseID); err != nil {
			return err
		}
		if err := database.Affected(tx.Model(&models.LoggedExercise{}).Scopes(own).
			Where("id = ?", exercise.ID).
			Updates(map[string]any{
				"workout_log_id": exercise.WorkoutLogID,
				"exercise_id":    exercise.ExerciseID,
				"notes":          exercise.Notes,
			})); err != nil {
			return err
		}

		incomingSetIDs := make(map[uint]struct{}, len(exercise.Sets))
		for i := range exercise.Sets {
			set := exercise.Sets[i]
			set.LoggedExerciseID = exercise.ID

			if set.ID > 0 {
				incomingSetIDs[set.ID] = struct{}{}
				if err := tx.Model(&models.LoggedSet{}).Scopes(own).
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
			if err := tx.Omit(clause.Associations).Create(&set).Error; err != nil {
				return err
			}
			incomingSetIDs[set.ID] = struct{}{}
		}

		var existingSetIDs []uint
		if err := tx.Model(&models.LoggedSet{}).Scopes(own).
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
			if err := tx.Unscoped().Scopes(own).
				Where("logged_exercise_id = ? AND id IN ?", exercise.ID, setIDsToDelete).
				Delete(&models.LoggedSet{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func RemoveLoggedExerciseForDay(ctx context.Context, day time.Time, exerciseID uint) error {
	res := database.From(ctx).WithContext(ctx).Unscoped().Scopes(database.Owned(ctx)).
		Where(
			"exercise_id = ? AND workout_log_id IN (?)",
			exerciseID,
			database.From(ctx).Model(&models.WorkoutLog{}).Scopes(database.Owned(ctx)).Select("id").Where("date = ?", day),
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

func LoadLoggedExercise(ctx context.Context, id uint) (models.LoggedExercise, error) {
	own := database.Owned(ctx)
	var exercise models.LoggedExercise
	err := database.From(ctx).Scopes(own).Preload("Exercise").
		Preload("Sets", func(db *gorm.DB) *gorm.DB { return setsInLoggedOrder(own(db)) }).
		Where("id = ?", id).First(&exercise).Error
	if err != nil {
		return models.LoggedExercise{}, err
	}
	if exercise.Exercise != nil {
		if err := DecorateExercises(ctx, exercise.Exercise); err != nil {
			return models.LoggedExercise{}, err
		}
	}
	if err := attachWorkoutLogDate(ctx, &exercise); err != nil {
		return models.LoggedExercise{}, err
	}
	return exercise, nil
}

func attachWorkoutLogDate(ctx context.Context, exerciseLog *models.LoggedExercise) error {
	if exerciseLog == nil || exerciseLog.ID == 0 {
		return nil
	}
	var workoutLog models.WorkoutLog
	err := database.From(ctx).Scopes(database.Owned(ctx)).Select("date").First(&workoutLog, exerciseLog.WorkoutLogID).Error
	if err != nil {
		return err
	}
	exerciseLog.LogDate = workoutLog.Date
	return nil
}

func GetPreviousExerciseLog(ctx context.Context, day time.Time, exercise string, offset int) (models.LoggedExercise, error) {
	own := database.Owned(ctx)
	uid, err := callerID(ctx)
	if err != nil {
		return models.LoggedExercise{}, err
	}
	var exerciseLog models.LoggedExercise
	err = database.From(ctx).WithContext(ctx).Scopes(own).
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id AND workout_logs.user_id = ?", uid).
		Joins("JOIN exercises ON exercises.id = logged_exercises.exercise_id").
		Where("exercises.name = ?", exercise).
		Where("workout_logs.date != ?", day).
		Where("workout_logs.date < ?", day).
		Preload("Sets", func(db *gorm.DB) *gorm.DB { return setsInLoggedOrder(own(db)) }).
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
	if err := attachWorkoutLogDate(ctx, &exerciseLog); err != nil {
		return models.LoggedExercise{}, err
	}
	return exerciseLog, nil
}

func GetMaxExerciseLog(ctx context.Context, day time.Time, exercise string) (models.LoggedExercise, error) {
	own := database.Owned(ctx)
	uid, err := callerID(ctx)
	if err != nil {
		return models.LoggedExercise{}, err
	}
	var exerciseLog models.LoggedExercise
	err = database.From(ctx).WithContext(ctx).Scopes(own).
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id AND workout_logs.user_id = ?", uid).
		Joins("JOIN exercises ON exercises.id = logged_exercises.exercise_id").
		Joins("JOIN logged_sets ON logged_sets.logged_exercise_id = logged_exercises.id AND logged_sets.user_id = ?", uid).
		Where("exercises.name = ?", exercise).
		Where("workout_logs.date != ?", day).
		Where("workout_logs.date < ?", day).
		Where("workout_logs.deleted_at IS NULL").
		Where("exercises.deleted_at IS NULL").
		Where("logged_sets.deleted_at IS NULL").
		Preload("Sets", func(db *gorm.DB) *gorm.DB { return setsInLoggedOrder(own(db)) }).
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
	if err := attachWorkoutLogDate(ctx, &exerciseLog); err != nil {
		return models.LoggedExercise{}, err
	}
	return exerciseLog, nil
}

// setsInLoggedOrder preloads sets in the order they were logged. Progression reads the final set as
// where the next session starts, and Postgres returns unordered rows in heap order, which an edit can change.
func setsInLoggedOrder(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }

func callerID(ctx context.Context) (uint, error) {
	id, ok := database.CurrentIdentity(ctx)
	if !ok {
		return 0, database.ErrNoIdentity
	}
	return id.UserID, nil
}
