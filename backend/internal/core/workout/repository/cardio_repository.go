package workoutrepo

import (
	"be-simpletracker/internal/database"
	"context"

	"be-simpletracker/internal/core/workout/models"

	"gorm.io/gorm/clause"
)

func FirstCardioByWorkoutLogID(ctx context.Context, workoutLogID uint) (models.Cardio, error) {
	var existing models.Cardio
	err := database.From(ctx).WithContext(ctx).Scopes(database.Owned(ctx)).Where("workout_log_id = ?", workoutLogID).First(&existing).Error
	return existing, err
}

func CreateCardio(ctx context.Context, row *models.Cardio) error {
	return database.From(ctx).WithContext(ctx).Omit(clause.Associations).Create(row).Error
}

// SaveCardio updates the three editable columns of the caller's cardio row. Not Save: that
// rewrites every column (user_id too) and falls back to an insert when the row is gone.
func SaveCardio(ctx context.Context, row *models.Cardio) error {
	return database.Affected(database.From(ctx).WithContext(ctx).Model(&models.Cardio{}).Scopes(database.Owned(ctx)).
		Where("id = ?", row.ID).
		Select("Minutes", "Type", "Notes", "UpdatedAt").
		Updates(row))
}
