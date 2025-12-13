package services

import (
	"be-simpletracker/database/models"
	"be-simpletracker/utils"
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

func LogExercise(db *gorm.DB, exercise models.LoggedExercise) error {
    err := db.Omit("Exercise").Create(&exercise).Error
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
    var plans []models.WorkoutPlan
    err := db.Preload("Exercises").Find(&plans).Error
    if err != nil {
        return []models.WorkoutPlan{}, err
    }
    return plans, nil
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