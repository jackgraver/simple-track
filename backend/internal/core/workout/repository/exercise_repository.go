package workoutrepo

import (
	"be-simpletracker/internal/database"
	"context"
	"time"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExerciseListResult struct {
	Exercises []models.Exercise
	Total     int64
}

type ExerciseProgressionEntry struct {
	Date   time.Time `json:"date"`
	Weight float32   `json:"weight"`
	Reps   uint      `json:"reps"`
	// Unilateral sets: each side is its own entry; the client pairs them by pair_id (or client_id).
	Side     *string `json:"side"`
	PairID   *string `json:"pair_id"`
	ClientID *string `json:"client_id"`
}

// DecorateExercises fills the caller-specific fields of shared exercises: CanEdit (the creator or
// the admin may change the shared row) and the caller's own machine settings, read in one query
// from exercise_settings. Exercises don't carry personal data themselves, so every read that goes
// out in a response passes through here.
func DecorateExercises(ctx context.Context, exercises ...*models.Exercise) error {
	if len(exercises) == 0 {
		return nil
	}
	id, ok := database.CurrentIdentity(ctx)
	if !ok {
		return database.ErrNoIdentity
	}
	ids := make([]uint, 0, len(exercises))
	for _, e := range exercises {
		ids = append(ids, e.ID)
	}
	var rows []models.ExerciseSetting
	if err := database.From(ctx).WithContext(ctx).Scopes(database.Owned(ctx)).
		Where("exercise_id IN ?", ids).Find(&rows).Error; err != nil {
		return err
	}
	byExercise := make(map[uint][]models.MachineSetting, len(rows))
	for _, r := range rows {
		byExercise[r.ExerciseID] = r.MachineSettings
	}
	for _, e := range exercises {
		e.MachineSettings = byExercise[e.ID]
		e.CanEdit = id.IsAdmin || e.CreatedBy == id.UserID
	}
	return nil
}

func decorateAll(ctx context.Context, exercises []models.Exercise) error {
	ptrs := make([]*models.Exercise, len(exercises))
	for i := range exercises {
		ptrs[i] = &exercises[i]
	}
	return DecorateExercises(ctx, ptrs...)
}

func FindAllExercises(ctx context.Context, excludeIDs []uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	query := database.From(ctx).Model(&models.Exercise{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	err := query.Find(&exercises).Error
	if err != nil {
		return []models.Exercise{}, err
	}
	if err := decorateAll(ctx, exercises); err != nil {
		return []models.Exercise{}, err
	}
	return exercises, nil
}

func ListExercises(ctx context.Context, page, pageSize int, search string) (ExerciseListResult, error) {
	query := database.From(ctx).Model(&models.Exercise{})
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	query = query.Order("name ASC")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ExerciseListResult{}, err
	}

	var exercises []models.Exercise
	if pageSize > 0 {
		query = query.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	if err := query.Find(&exercises).Error; err != nil {
		return ExerciseListResult{}, err
	}
	if err := decorateAll(ctx, exercises); err != nil {
		return ExerciseListResult{}, err
	}

	return ExerciseListResult{Exercises: exercises, Total: total}, nil
}

// GetExerciseProgression is the caller's sets for one exercise: every table in the join is
// filtered by the caller (RLS does the same underneath).
func GetExerciseProgression(ctx context.Context, exerciseID uint) ([]ExerciseProgressionEntry, error) {
	id, ok := database.CurrentIdentity(ctx)
	if !ok {
		return nil, database.ErrNoIdentity
	}
	var entries []ExerciseProgressionEntry

	err := database.From(ctx).
		Table("logged_exercises").
		Select("workout_logs.date, logged_sets.weight, logged_sets.reps, logged_sets.side, logged_sets.pair_id, logged_sets.client_id").
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id AND workout_logs.user_id = ?", id.UserID).
		Joins("JOIN logged_sets ON logged_sets.logged_exercise_id = logged_exercises.id AND logged_sets.user_id = ?", id.UserID).
		Where("logged_exercises.user_id = ?", id.UserID).
		Where("logged_exercises.exercise_id = ?", exerciseID).
		Where("logged_sets.weight > 0 AND logged_sets.reps > 0").
		Order("workout_logs.date ASC, logged_sets.id ASC").
		Scan(&entries).Error

	if err != nil {
		return []ExerciseProgressionEntry{}, err
	}
	return entries, nil
}

// ExerciseExists is nil when the (live) catalog exercise exists.
func ExerciseExists(ctx context.Context, id uint) error {
	return database.From(ctx).First(&models.Exercise{}, id).Error
}

// CreateExercise adds a shared exercise; the Creator hook stamps created_by with the caller.
func CreateExercise(ctx context.Context, exercise *models.Exercise) error {
	if err := database.From(ctx).Omit(clause.Associations).Create(exercise).Error; err != nil {
		return err
	}
	return DecorateExercises(ctx, exercise)
}

// loadEditable returns the live exercise when the caller may change it. A missing exercise is
// ErrNotFound; one the caller can see but didn't create (and isn't admin for) is ErrForbidden.
func loadEditable(ctx context.Context, tx *gorm.DB, id uint) (models.Exercise, error) {
	who, ok := database.CurrentIdentity(ctx)
	if !ok {
		return models.Exercise{}, database.ErrNoIdentity
	}
	var exercise models.Exercise
	if err := tx.First(&exercise, id).Error; err != nil {
		return models.Exercise{}, err
	}
	if !who.IsAdmin && exercise.CreatedBy != who.UserID {
		return models.Exercise{}, ErrForbidden
	}
	return exercise, nil
}

func UpdateExercise(ctx context.Context, id uint, name string, repRollover uint, cues string, loadTypes ...models.ExerciseLoadType) (*models.Exercise, error) {
	return UpdateExerciseWithUnilateral(ctx, id, name, repRollover, cues, nil, loadTypes...)
}

// UpdateExerciseWithUnilateral is UpdateExercise that can also turn one-side-at-a-time logging on
// or off (nil keeps it), in the same single write. Only the creator or the admin may edit.
func UpdateExerciseWithUnilateral(ctx context.Context, id uint, name string, repRollover uint, cues string, unilateral *bool, loadTypes ...models.ExerciseLoadType) (*models.Exercise, error) {
	db := database.From(ctx).WithContext(ctx)
	exercise, err := loadEditable(ctx, db, id)
	if err != nil {
		return nil, err
	}
	loadType := models.NormalizeExerciseLoadType(exercise.LoadType)
	if len(loadTypes) > 0 {
		loadType = models.NormalizeExerciseLoadType(loadTypes[0])
	}
	updates := map[string]interface{}{
		"name":         name,
		"rep_rollover": repRollover,
		"cues":         cues,
		"load_type":    loadType,
	}
	if unilateral != nil {
		updates["unilateral"] = *unilateral
	}
	if err := database.Affected(db.Model(&models.Exercise{}).Where("id = ?", id).Updates(updates)); err != nil {
		return nil, err
	}
	return reloadExercise(ctx, id)
}

func UpdateExerciseCues(ctx context.Context, exerciseID uint, cues string) (*models.Exercise, error) {
	db := database.From(ctx).WithContext(ctx)
	if _, err := loadEditable(ctx, db, exerciseID); err != nil {
		return nil, err
	}
	if err := database.Affected(db.Model(&models.Exercise{}).Where("id = ?", exerciseID).Update("cues", cues)); err != nil {
		return nil, err
	}
	return reloadExercise(ctx, exerciseID)
}

func reloadExercise(ctx context.Context, id uint) (*models.Exercise, error) {
	var exercise models.Exercise
	if err := database.From(ctx).WithContext(ctx).First(&exercise, id).Error; err != nil {
		return nil, err
	}
	if err := DecorateExercises(ctx, &exercise); err != nil {
		return nil, err
	}
	return &exercise, nil
}

// DeleteExercise soft-deletes a shared exercise. Admin only (the catalog_guard trigger enforces
// it too). It refuses with ErrExerciseInUse while a plan or log of the admin's still uses it. Other
// users' plans and logs are invisible to the admin (RLS), so their use can't be counted: deleting
// an exercise somebody else still uses hides it from their old logs. API only, no UI (DECISIONS).
func DeleteExercise(ctx context.Context, id uint) error {
	who, ok := database.CurrentIdentity(ctx)
	if !ok {
		return database.ErrNoIdentity
	}
	db := database.From(ctx).WithContext(ctx)
	if err := db.First(&models.Exercise{}, id).Error; err != nil {
		return err
	}
	if !who.IsAdmin {
		return ErrForbidden
	}
	var uses int64
	if err := db.Model(&models.WorkoutPlanExercise{}).Scopes(database.Owned(ctx)).
		Where("exercise_id = ?", id).Count(&uses).Error; err != nil {
		return err
	}
	if uses == 0 {
		if err := db.Model(&models.LoggedExercise{}).Scopes(database.Owned(ctx)).
			Where("exercise_id = ?", id).Count(&uses).Error; err != nil {
			return err
		}
	}
	if uses > 0 {
		return ErrExerciseInUse
	}
	return database.Affected(db.Delete(&models.Exercise{}, id))
}

// SaveMachineSettings stores the caller's machine settings for an exercise (an empty list clears
// them) and returns the exercise as the caller sees it. Any signed-in user may do this for any
// catalog exercise: the data is theirs, not the exercise's.
func SaveMachineSettings(ctx context.Context, exerciseID uint, settings []models.MachineSetting) (*models.Exercise, error) {
	db := database.From(ctx).WithContext(ctx)
	if err := db.First(&models.Exercise{}, exerciseID).Error; err != nil {
		return nil, err
	}
	own := database.Owned(ctx)
	err := db.Transaction(func(tx *gorm.DB) error {
		if len(settings) == 0 {
			return tx.Scopes(own).Where("exercise_id = ?", exerciseID).Delete(&models.ExerciseSetting{}).Error
		}
		var row models.ExerciseSetting
		err := tx.Scopes(own).Where("exercise_id = ?", exerciseID).First(&row).Error
		switch {
		case err == nil:
			return database.Affected(tx.Model(&models.ExerciseSetting{}).Scopes(own).Where("id = ?", row.ID).
				Select("MachineSettings", "UpdatedAt").
				Updates(&models.ExerciseSetting{MachineSettings: settings}))
		case database.IsNotFound(err):
			return tx.Omit(clause.Associations).Create(&models.ExerciseSetting{ExerciseID: exerciseID, MachineSettings: settings}).Error
		default:
			return err
		}
	})
	if err != nil {
		return nil, err
	}
	return reloadExercise(ctx, exerciseID)
}

// ExerciseRowExists reports whether an exercise row exists, including soft-deleted ones (old
// sessions may still log against them).
func ExerciseRowExists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := database.From(ctx).WithContext(ctx).Unscoped().Model(&models.Exercise{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
