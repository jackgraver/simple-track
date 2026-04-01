package services

import (
	"be-simpletracker/internal/features/workout/models"
	"be-simpletracker/internal/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func loadWorkoutLogForDate(database *gorm.DB, date time.Time) (models.WorkoutLog, error) {
	var workoutDay models.WorkoutLog
	err := database.
		Preload("Cardio").
		Preload("Exercises.Sets").
		Preload("Exercises.Exercise").
		Preload("WorkoutPlan.Exercises").
		Where("date = ?", date).
		First(&workoutDay).Error
	if err != nil {
		return models.WorkoutLog{}, err
	}
	return workoutDay, nil
}

func GetToday(database *gorm.DB, offset int) (models.WorkoutLog, error) {
	return loadWorkoutLogForDate(database, utils.ZerodTime(offset))
}

// GetOrCreateToday returns the workout log for the calendar day (with offset from today),
// creating a row if missing and attaching the plan for that weekday when one exists.
func GetOrCreateToday(database *gorm.DB, offset int) (models.WorkoutLog, error) {
	day := utils.ZerodTime(offset)
	workoutDay, err := loadWorkoutLogForDate(database, day)
	if err == nil {
		return workoutDay, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.WorkoutLog{}, err
	}
	plan, err := GetPlanByDay(database, int(day.Weekday()))
	if err != nil {
		return models.WorkoutLog{}, err
	}
	var planID *uint
	if plan != nil {
		id := plan.ID
		planID = &id
	}
	newLog := models.WorkoutLog{
		Date:          day,
		WorkoutPlanID: planID,
	}
	if err := database.Omit("WorkoutPlan", "Exercises", "Cardio").Create(&newLog).Error; err != nil {
		return models.WorkoutLog{}, err
	}
	return loadWorkoutLogForDate(database, day)
}

func GetAll(database *gorm.DB) ([]models.WorkoutLog, error) {
	var workoutDay []models.WorkoutLog

	err := database.
		Preload("Cardio").
		Preload("Exercises.Sets").
		Find(&workoutDay).Error

	if err != nil {
		return []models.WorkoutLog{}, err
	}
	return workoutDay, nil
}

func GetPrevious(db *gorm.DB, day string) (models.WorkoutLog, error) {
	var workoutDay models.WorkoutLog

	err := db.
		Preload("Cardio").
		Preload("Exercises.Sets").
		Where("WorkoutPlan.name = ?", day).
		Where("date < ?", time.Now()).
		Order("date DESC").
		Limit(1).
		Find(&workoutDay).Error

	if err != nil {
		return models.WorkoutLog{}, err
	}
	return workoutDay, nil
}

func GetPreviousExerciseLog(db *gorm.DB, day time.Time, exercise string, offset int) (models.LoggedExercise, error) {
	var exerciseLog models.LoggedExercise

	err := db.
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id").
		Joins("JOIN exercises ON exercises.id = logged_exercises.exercise_id").
		Where("exercises.name = ?", exercise).
		Where("workout_logs.date != ?", day).
		Where("workout_logs.date < ?", day).
		Preload("Sets").
		Preload("Exercise").
		Order("workout_logs.date DESC").
		Offset(offset).
		Limit(1).
		Find(&exerciseLog).Error

	if err != nil {
		return models.LoggedExercise{}, err
	}
	return exerciseLog, nil
}

func LogExercise(db *gorm.DB, exercise *models.LoggedExercise) error {
	err := db.Omit("Exercise").Create(exercise).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateLoggedExercise(db *gorm.DB, exercise models.LoggedExercise) error {
	return db.Transaction(func(tx *gorm.DB) error {
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

func DeleteLoggedSet(db *gorm.DB, setID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
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

func GetAllExercises(db *gorm.DB, excludeIDs []uint) ([]models.Exercise, error) {
	var exercises []models.Exercise
	query := db.Model(&models.Exercise{})
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	err := query.Find(&exercises).Error
	if err != nil {
		return []models.Exercise{}, err
	}
	return exercises, nil
}

type ExerciseProgressionEntry struct {
	Date   time.Time `json:"date"`
	Weight float32   `json:"weight"`
	Reps   uint      `json:"reps"`
}

func GetExerciseProgression(db *gorm.DB, exerciseID uint) ([]ExerciseProgressionEntry, error) {
	var entries []ExerciseProgressionEntry

	err := db.
		Table("logged_exercises").
		Select("workout_logs.date, logged_sets.weight, logged_sets.reps").
		Joins("JOIN workout_logs ON workout_logs.id = logged_exercises.workout_log_id").
		Joins("JOIN logged_sets ON logged_sets.logged_exercise_id = logged_exercises.id").
		Where("logged_exercises.exercise_id = ?", exerciseID).
		Where("logged_sets.weight > 0 AND logged_sets.reps > 0").
		Order("workout_logs.date ASC").
		Scan(&entries).Error

	if err != nil {
		return []ExerciseProgressionEntry{}, err
	}
	return entries, nil
}

func GetAllWorkoutPlans(db *gorm.DB) ([]models.WorkoutPlan, error) {
	var workoutPlans []models.WorkoutPlan
	err := db.Preload("Exercises").Find(&workoutPlans).Error
	if err != nil {
		return []models.WorkoutPlan{}, err
	}
	return workoutPlans, nil
}

func AddExerciseToPlan(db *gorm.DB, planID uint, exerciseID uint) error {
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return err
	}

	var exercise models.Exercise
	if err := db.First(&exercise, exerciseID).Error; err != nil {
		return err
	}

	return db.Model(&plan).Association("Exercises").Append(&exercise)
}

func RemoveExerciseFromPlan(db *gorm.DB, planID uint, exerciseID uint) error {
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return err
	}

	var exercise models.Exercise
	if err := db.First(&exercise, exerciseID).Error; err != nil {
		return err
	}

	return db.Model(&plan).Association("Exercises").Delete(&exercise)
}

func CreateExercise(db *gorm.DB, name string, repRollover uint) (*models.Exercise, error) {
	exercise := models.Exercise{
		Name:        name,
		RepRollover: repRollover,
	}

	if err := db.Create(&exercise).Error; err != nil {
		return nil, err
	}

	return &exercise, nil
}

// AssignPlanToDay assigns a workout plan to a specific day of the week
// If another plan is already assigned to that day, it will be unassigned first
// dayOfWeek: 0=Sunday, 1=Monday, ..., 6=Saturday
func AssignPlanToDay(db *gorm.DB, planID uint, dayOfWeek int) (*models.WorkoutPlan, error) {
	// Validate dayOfWeek
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}

	// Find the plan
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Unassign any existing plan from this day
	if err := db.Model(&models.WorkoutPlan{}).
		Where("day_of_week = ? AND id != ?", dayOfWeek, planID).
		Update("day_of_week", nil).Error; err != nil {
		return nil, fmt.Errorf("failed to unassign existing plan: %w", err)
	}

	// Assign the plan to the day
	dayOfWeekPtr := &dayOfWeek
	if err := db.Model(&plan).Update("day_of_week", dayOfWeekPtr).Error; err != nil {
		return nil, fmt.Errorf("failed to assign plan to day: %w", err)
	}

	// Reload the plan with exercises
	if err := db.Preload("Exercises").First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload plan: %w", err)
	}

	return &plan, nil
}

// UnassignPlanFromDay removes the day assignment from a workout plan
func UnassignPlanFromDay(db *gorm.DB, planID uint) (*models.WorkoutPlan, error) {
	var plan models.WorkoutPlan
	if err := db.First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	if err := db.Model(&plan).Update("day_of_week", nil).Error; err != nil {
		return nil, fmt.Errorf("failed to unassign plan from day: %w", err)
	}

	// Reload the plan with exercises
	if err := db.Preload("Exercises").First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload plan: %w", err)
	}

	return &plan, nil
}

// GetPlanByDay returns the workout plan assigned to a specific day of the week
// UpsertCardioForWorkoutLog creates or updates the single cardio row for a workout log.
func UpsertCardioForWorkoutLog(db *gorm.DB, offset int, minutes int, cardioType string, notes string) (*models.Cardio, error) {
	t, err := GetOrCreateToday(db, offset)
	if err != nil {
		return nil, err
	}
	ctype := strings.TrimSpace(cardioType)
	if ctype == "" && t.WorkoutPlan != nil {
		ctype = strings.TrimSpace(t.WorkoutPlan.PlannedCardioType)
	}
	if ctype == "" {
		return nil, fmt.Errorf("cardio type is required when the plan has no planned cardio")
	}
	var existing models.Cardio
	err = db.Where("workout_log_id = ?", t.ID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := models.Cardio{
			WorkoutLogID: t.ID,
			Minutes:      minutes,
			Type:         ctype,
			Notes:        notes,
		}
		if err := db.Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	existing.Minutes = minutes
	existing.Type = ctype
	existing.Notes = notes
	if err := db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func GetPlanByDay(db *gorm.DB, dayOfWeek int) (*models.WorkoutPlan, error) {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}

	var plan models.WorkoutPlan
	err := db.Preload("Exercises").Where("day_of_week = ?", dayOfWeek).First(&plan).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No plan assigned to this day
		}
		return nil, err
	}

	return &plan, nil
}
