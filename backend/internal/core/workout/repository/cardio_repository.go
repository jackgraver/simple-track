package workoutrepo

import (
	"context"

	"be-simpletracker/internal/core/workout/models"
)

func FirstCardioByWorkoutLogID(ctx context.Context, workoutLogID uint) (models.Cardio, error) {
	var existing models.Cardio
	err := conn().WithContext(ctx).Where("workout_log_id = ?", workoutLogID).First(&existing).Error
	return existing, err
}

func CreateCardio(ctx context.Context, row *models.Cardio) error {
	return conn().WithContext(ctx).Create(row).Error
}

func SaveCardio(ctx context.Context, row *models.Cardio) error {
	return conn().WithContext(ctx).Save(row).Error
}
