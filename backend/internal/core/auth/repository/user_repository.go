package authrepo

import (
	"context"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

func FindUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User
	if err := database.From(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindUserByID returns a live (not soft-deleted) user. Disabled users are returned; callers
// decide what disabled means for them.
func FindUserByID(ctx context.Context, id uint) (models.User, error) {
	var user models.User
	if err := database.From(ctx).First(&user, id).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func CreateUser(ctx context.Context, user *models.User) error {
	if user.TokenVersion == 0 {
		user.TokenVersion = models.InitialTokenVersion
	}
	return database.From(ctx).Create(user).Error
}

func UpdateUserBirthYear(ctx context.Context, id uint, birthYear *int) (models.User, error) {
	result := database.From(ctx).Model(&models.User{}).Where("id = ?", id).Update("birth_year", birthYear)
	if result.Error != nil {
		return models.User{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return FindUserByID(ctx, id)
}
