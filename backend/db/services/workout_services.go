package services

import (
	"be-simpletracker/db/models"
	"time"

	"gorm.io/gorm"
)

func GetToday(database *gorm.DB) (models.WorkoutLog, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	var workoutDay models.WorkoutLog
	err := database.
		Preload("WorkoutPlan.Exercises.Sets").
		Preload("Cardio").
		Preload("Exercises.Sets").
		Where("date >= ? AND date < ?", start, end).
		First(&workoutDay).Error

	if err != nil {
		return models.WorkoutLog{}, err
	}
	return workoutDay, nil
}

func GetAll(database *gorm.DB) ([]models.WorkoutLog, error) {
    var workoutDay []models.WorkoutLog

    err := database.
        Preload("WorkoutPlan.Exercises.Sets").
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
        Joins("WorkoutPlan").
        Preload("WorkoutPlan.Exercises.Sets").
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