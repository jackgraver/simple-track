package profile

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var allowedSex = map[string]struct{}{
	"male":   {},
	"female": {},
}

var allowedActivity = map[string]struct{}{
	"sedentary":         {},
	"lightly_active":    {},
	"moderately_active": {},
	"very_active":       {},
	"extra_active":      {},
}

func ValidateSex(s string) error {
	if _, ok := allowedSex[strings.ToLower(strings.TrimSpace(s))]; !ok {
		return errors.New("sex must be male or female")
	}
	return nil
}

func ValidateActivity(a string) error {
	if _, ok := allowedActivity[strings.TrimSpace(a)]; !ok {
		return errors.New("invalid activity_level")
	}
	return nil
}

func GetProfile(db *gorm.DB) (*UserProfile, error) {
	var row UserProfile
	err := db.Order("id ASC").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func UpsertProfile(db *gorm.DB, heightIn float64, age int, sex, activityLevel string) (*UserProfile, error) {
	if heightIn <= 0 {
		return nil, errors.New("height_in must be positive")
	}
	if age <= 0 || age > 130 {
		return nil, errors.New("age must be between 1 and 130")
	}
	sex = strings.ToLower(strings.TrimSpace(sex))
	if err := ValidateSex(sex); err != nil {
		return nil, err
	}
	activityLevel = strings.TrimSpace(activityLevel)
	if err := ValidateActivity(activityLevel); err != nil {
		return nil, err
	}
	var row UserProfile
	err := db.Order("id ASC").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = UserProfile{
			HeightIn:      heightIn,
			Age:           age,
			Sex:           sex,
			ActivityLevel: activityLevel,
		}
		if err := db.Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	row.HeightIn = heightIn
	row.Age = age
	row.Sex = sex
	row.ActivityLevel = activityLevel
	if err := db.Save(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
