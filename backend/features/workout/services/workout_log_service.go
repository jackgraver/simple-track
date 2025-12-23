package services

import (
	"be-simpletracker/features/workout/models"
	"be-simpletracker/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func GetToday(database *gorm.DB, offset int) (models.WorkoutLog, error) {
	today := utils.ZerodTime(offset)

	var workoutDay models.WorkoutLog
	err := database.
		Preload("Cardio").
		Preload("Exercises.Sets").
		Preload("Exercises.Exercise").
		Preload("WorkoutPlan.Exercises").
		Where("date = ?", today).
		First(&workoutDay).Error

	if err != nil {
		return models.WorkoutLog{}, err
	}
	return workoutDay, nil
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
	err := db.Omit("Exercise").Updates(&exercise).Error
	if err != nil {
		return err
	}
	return nil
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