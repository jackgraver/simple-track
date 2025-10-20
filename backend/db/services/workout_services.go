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
		Preload("WorkoutPlan.Exercises.Sets").
		Preload("Cardio").
		Preload("Exercises.Sets").
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