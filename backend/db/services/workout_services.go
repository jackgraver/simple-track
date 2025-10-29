package services

import (
	"be-simpletracker/db/models"
	"be-simpletracker/utils"
	"time"

	"gorm.io/gorm"
)

func GetToday(database *gorm.DB) (models.WorkoutLog, error) {
	today := utils.ZerodTime()

	var workoutDay models.WorkoutLog
	err := database.
		Preload("Cardio").
		Preload("Exercises.Sets").
        Preload("Exercises.Exercise").
        Preload("WorkoutPlan").
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
    err := db.Create(&exercise).Error
    if err != nil {
        return err
    }
    return nil
}