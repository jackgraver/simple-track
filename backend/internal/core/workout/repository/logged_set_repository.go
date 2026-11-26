package workoutrepo

import (
	"context"
	"errors"
	"time"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	res, err := DeleteSet(context.Background(), SetRef{ID: setID})
	if err != nil {
		return err
	}
	if !res.Deleted {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ErrClientIDConflict means a set with this client_id already exists under a different
// workout day or exercise, so the request is not a replay of the original create.
var ErrClientIDConflict = errors.New("client_id already used for a different set")

// SetRef identifies a set by server id or by its client_id (exactly one is set).
type SetRef struct {
	ID       uint
	ClientID string
}

func (r SetRef) where(tx *gorm.DB) *gorm.DB {
	if r.ClientID != "" {
		return tx.Where("client_id = ?", r.ClientID)
	}
	return tx.Where("id = ?", r.ID)
}

type CreateSetInput struct {
	ClientID     string
	WorkoutLogID uint
	ExerciseID   uint
	Reps         uint
	Weight       float32
	WeightSetup  string
	Side         *string
}

type CreateSetResult struct {
	Set              models.LoggedSet
	LoggedExerciseID uint
	// Created is false when the request replayed an earlier create with the same client_id.
	Created bool
	// FirstSetOfDay is true when no other committed set existed for the workout log. Racing
	// first creates may each report true; at least one always does.
	FirstSetOfDay bool
}

// CreateSet logs one set idempotently. It upserts the parent logged_exercise (unique per workout
// log and exercise), then inserts the set with ON CONFLICT (client_id) DO NOTHING. A replay with
// the same client_id returns the stored set unchanged (first write wins).
func CreateSet(ctx context.Context, in CreateSetInput) (CreateSetResult, error) {
	var out CreateSetResult
	err := conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if existing, ok, err := findReplay(tx, in); err != nil || ok {
			out = existing
			return err
		}

		parentID, err := upsertAndLockParent(tx, in.WorkoutLogID, in.ExerciseID)
		if err != nil {
			return err
		}
		clientID := in.ClientID
		set := models.LoggedSet{
			LoggedExerciseID: parentID,
			ClientID:         &clientID,
			Reps:             in.Reps,
			Weight:           in.Weight,
			WeightSetup:      in.WeightSetup,
			Side:             in.Side,
		}
		res := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "client_id"}}, DoNothing: true}).Create(&set)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 || set.ID == 0 {
			// A concurrent replay committed first.
			existing, ok, err := findReplay(tx, in)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("set insert skipped but no row with that client_id")
			}
			out = existing
			return nil
		}

		var daySets int64
		if err := tx.Model(&models.LoggedSet{}).
			Joins("JOIN logged_exercises ON logged_exercises.id = logged_sets.logged_exercise_id").
			Where("logged_exercises.workout_log_id = ? AND logged_exercises.deleted_at IS NULL", in.WorkoutLogID).
			Count(&daySets).Error; err != nil {
			return err
		}
		out = CreateSetResult{Set: set, LoggedExerciseID: parentID, Created: true, FirstSetOfDay: daySets == 1}
		return nil
	})
	return out, err
}

// upsertAndLockParent returns the live logged_exercise for (workout log, exercise), creating it
// if needed, and holds a KEY SHARE lock on it until the transaction ends. DeleteSet takes FOR
// UPDATE on the parent before deciding to delete it, so the two serialize: either the delete sees
// this transaction's set and keeps the parent, or it deletes the parent first and the lock here
// finds no row, in which case the upsert runs again and creates a fresh parent. Without the lock
// a racing delete of the last set would cascade away a set this transaction already committed.
func upsertAndLockParent(tx *gorm.DB, workoutLogID, exerciseID uint) (uint, error) {
	for attempt := 0; attempt < 5; attempt++ {
		parentID, _, err := upsertLoggedExercise(tx, workoutLogID, exerciseID, "")
		if err != nil {
			return 0, err
		}
		var locked []uint
		if err := tx.Model(&models.LoggedExercise{}).
			Clauses(clause.Locking{Strength: "KEY SHARE"}).
			Where("id = ?", parentID).
			Pluck("id", &locked).Error; err != nil {
			return 0, err
		}
		if len(locked) == 1 {
			return parentID, nil
		}
	}
	return 0, errors.New("logged exercise kept disappearing while logging a set")
}

// findReplay returns the set already stored under in.ClientID, if any. It fails with
// ErrClientIDConflict when that set belongs to another workout log or exercise.
func findReplay(tx *gorm.DB, in CreateSetInput) (CreateSetResult, bool, error) {
	var set models.LoggedSet
	if err := tx.Where("client_id = ?", in.ClientID).Limit(1).Find(&set).Error; err != nil {
		return CreateSetResult{}, false, err
	}
	if set.ID == 0 {
		return CreateSetResult{}, false, nil
	}
	var parent models.LoggedExercise
	if err := tx.Unscoped().Select("id", "workout_log_id", "exercise_id").
		First(&parent, set.LoggedExerciseID).Error; err != nil {
		return CreateSetResult{}, false, err
	}
	if parent.WorkoutLogID != in.WorkoutLogID || parent.ExerciseID != in.ExerciseID {
		return CreateSetResult{}, false, ErrClientIDConflict
	}
	return CreateSetResult{Set: set, LoggedExerciseID: parent.ID}, true, nil
}

// SetPatch holds the fields to change; nil fields are left alone.
type SetPatch struct {
	Reps        *uint
	Weight      *float32
	WeightSetup *string
	// Side "" stores NULL (bilateral); "left"/"right" set the side. Validated by the caller.
	Side *string
}

// UpdateSet applies patch to one set and returns it. gorm.ErrRecordNotFound if it doesn't exist.
func UpdateSet(ctx context.Context, ref SetRef, patch SetPatch) (models.LoggedSet, error) {
	var set models.LoggedSet
	err := conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ref.where(tx).First(&set).Error; err != nil {
			return err
		}
		updates := map[string]any{}
		if patch.Reps != nil {
			updates["reps"] = *patch.Reps
		}
		if patch.Weight != nil {
			updates["weight"] = *patch.Weight
		}
		if patch.WeightSetup != nil {
			updates["weight_setup"] = *patch.WeightSetup
		}
		if patch.Side != nil {
			side, _ := models.NormalizeSide(patch.Side)
			updates["side"] = side
		}
		if len(updates) == 0 {
			return nil
		}
		if err := tx.Model(&models.LoggedSet{}).Where("id = ?", set.ID).Updates(updates).Error; err != nil {
			return err
		}
		return tx.First(&set, set.ID).Error
	})
	return set, err
}

type DeleteSetResult struct {
	// Deleted is false when the set was already gone (a repeat delete).
	Deleted          bool
	LoggedExerciseID uint
	// ExerciseRemoved is true when this was the exercise's last set, so its logged_exercise
	// row was deleted too (the legacy behaviour, kept).
	ExerciseRemoved bool
}

// DeleteSet hard-deletes one set, and its logged_exercise when no sets remain. A missing set is
// not an error: Deleted is false.
func DeleteSet(ctx context.Context, ref SetRef) (DeleteSetResult, error) {
	var out DeleteSetResult
	err := conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var set models.LoggedSet
		if err := ref.where(tx).Limit(1).Find(&set).Error; err != nil {
			return err
		}
		if set.ID == 0 {
			return nil
		}
		// Lock the parent before counting its sets, so a concurrent CreateSet for the same
		// exercise either commits its set first (and the count sees it) or waits for this
		// delete and then creates a new parent (see upsertAndLockParent).
		var parentIDs []uint
		if err := tx.Unscoped().Model(&models.LoggedExercise{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", set.LoggedExerciseID).
			Pluck("id", &parentIDs).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&models.LoggedSet{}, set.ID).Error; err != nil {
			return err
		}
		out = DeleteSetResult{Deleted: true, LoggedExerciseID: set.LoggedExerciseID}

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
			out.ExerciseRemoved = true
		}
		return nil
	})
	return out, err
}
