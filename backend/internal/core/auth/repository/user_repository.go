package authrepo

import "be-simpletracker/internal/core/auth/models"

func FindUserByUsername(username string) (models.User, error) {
	var user models.User
	if err := conn().Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// FindUserByID returns a live (not soft-deleted) user. Disabled users are returned; callers
// decide what disabled means for them.
func FindUserByID(id uint) (models.User, error) {
	var user models.User
	if err := conn().First(&user, id).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func CreateUser(user *models.User) error {
	if user.TokenVersion == 0 {
		user.TokenVersion = models.InitialTokenVersion
	}
	return conn().Create(user).Error
}

func UpdateUserBirthYear(id uint, birthYear *int) (models.User, error) {
	if err := conn().Model(&models.User{}).Where("id = ?", id).Update("birth_year", birthYear).Error; err != nil {
		return models.User{}, err
	}
	return FindUserByID(id)
}
