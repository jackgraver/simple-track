package workout

import (
	"time"

	"gorm.io/gorm"
)

func GetToday(database *gorm.DB) (WorkoutLog, error) {
    today := time.Now().Truncate(24 * time.Hour)

    var workoutDay WorkoutLog
    err := database.
        Preload("WorkoutPlan.Exercises.Sets").
        Preload("Cardio").
        Preload("Exercises.Sets").
        Where("date = ?", today).
        First(&workoutDay).Error

    if err != nil {
        return WorkoutLog{}, err
    }
    return workoutDay, nil
}

func GetAll(database *gorm.DB) ([]WorkoutLog, error) {
    var workoutDay []WorkoutLog

    err := database.
        Preload("WorkoutPlan.Exercises.Sets").
        Preload("Cardio").
        Preload("Exercises.Sets").
        Find(&workoutDay).Error

    if err != nil {
        return []WorkoutLog{}, err
    }
    return workoutDay, nil
}
