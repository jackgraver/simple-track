package workout

import (
	"time"

	"gorm.io/gorm"
)

// NOTE: No database calls yet. Keep functions simple and return TODO errors.

func GetToday(database *gorm.DB) (WorkoutDay, error) {
    today := time.Now().Truncate(24 * time.Hour)
    var workoutDay WorkoutDay
    if err := database.
            Preload("Exercises").
            Preload("Cardio").
            First(&workoutDay, today).Error; err != nil {
        return WorkoutDay{}, err
    }

    return WorkoutDay{}, nil
}

