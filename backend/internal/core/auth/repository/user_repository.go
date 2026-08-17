package authrepo

import "be-simpletracker/internal/core/auth/models"

func FindUserByUsername(username string) (models.User, error) {
	var user models.User
	if err := conn().Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func CreateUser(user *models.User) error {
	return conn().Create(user).Error
}

func UpdateUserBirthYear(username string, birthYear *int) (models.User, error) {
	var user models.User
	if err := conn().Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, err
	}
	if err := conn().Model(&user).Update("birth_year", birthYear).Error; err != nil {
		return models.User{}, err
	}
	if err := conn().Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}
