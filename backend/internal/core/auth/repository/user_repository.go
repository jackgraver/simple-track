package authrepo

import (
	"context"
	"errors"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// FindUserByUsername matches case-insensitively (usernames are unique on lower(username)).
func FindUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User
	if err := database.From(ctx).Where("lower(username) = lower(?)", username).First(&user).Error; err != nil {
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

// ErrUsernameTaken: the insert hit the case-insensitive username index (a soft-deleted user with
// that name, or a concurrent create).
var ErrUsernameTaken = errors.New("username already taken")

func CreateUser(ctx context.Context, user *models.User) error {
	if user.TokenVersion == 0 {
		user.TokenVersion = models.InitialTokenVersion
	}
	db := database.From(ctx)
	if user.Email == "" {
		// NULL, not '': accounts without an email don't collide on the email unique index.
		db = db.Omit("Email")
	}
	// In a savepoint: the request can still answer after the unique violation (B2 gotcha).
	err := database.CreateInSavepoint(db, user)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "idx_users_username_lower" {
		return ErrUsernameTaken
	}
	return err
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
