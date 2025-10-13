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

func WorkoutRange(db *gorm.DB, today time.Time, start time.Time, end time.Time) ([]WorkoutLog, error) {
    var logs []WorkoutLog

	if err := db.
		Preload("WorkoutPlan.Exercises.Sets").
		Preload("Cardio").
		Preload("Exercises.Sets").
		Where("date BETWEEN ? AND ?", start, end).
		Order("date").
		Find(&logs).Error; err != nil {
		return nil, err
	}
    return logs, nil
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
