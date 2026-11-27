package workoutrepo

import (
	"be-simpletracker/internal/database"
	"context"

	"be-simpletracker/internal/core/workout/models"
)

func FirstCardioByWorkoutLogID(ctx context.Context, workoutLogID uint) (models.Cardio, error) {
	var existing models.Cardio
	err := database.From(ctx).WithContext(ctx).Where("workout_log_id = ?", workoutLogID).First(&existing).Error
	return existing, err
}

func CreateCardio(ctx context.Context, row *models.Cardio) error {
	return database.From(ctx).WithContext(ctx).Create(row).Error
}

func SaveCardio(ctx context.Context, row *models.Cardio) error {
	return database.From(ctx).WithContext(ctx).Save(row).Error
}
