package profile

import (
	"be-simpletracker/internal/core/tracking/common"
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		return common.Invalid("sex must be male or female")
	}
	return nil
}

func ValidateActivity(a string) error {
	if _, ok := allowedActivity[strings.TrimSpace(a)]; !ok {
		return common.Invalid("invalid activity_level")
	}
	return nil
}

// GetProfile returns the current user's profile, or nil when they haven't saved one yet (a
// profile is created on first save, so new accounts start without one).
func GetProfile(ctx context.Context) (*UserProfile, error) {
	var row UserProfile
	err := database.From(ctx).Scopes(database.Owned(ctx)).Order("id ASC").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpsertProfile saves the current user's profile, creating it on first use. There is one live row
// per user (unique index, migration 0013).
func UpsertProfile(ctx context.Context, heightIn float64, age int, sex, activityLevel string) (*UserProfile, error) {
	if heightIn <= 0 {
		return nil, common.Invalid("height_in must be positive")
	}
	if age <= 0 || age > 130 {
		return nil, common.Invalid("age must be between 1 and 130")
	}
	sex = strings.ToLower(strings.TrimSpace(sex))
	if err := ValidateSex(sex); err != nil {
		return nil, err
	}
	activityLevel = strings.TrimSpace(activityLevel)
	if err := ValidateActivity(activityLevel); err != nil {
		return nil, err
	}
	db := database.From(ctx)
	update := func(row *UserProfile) (*UserProfile, error) {
		row.HeightIn = heightIn
		row.Age = age
		row.Sex = sex
		row.ActivityLevel = activityLevel
		if err := db.Omit(clause.Associations).Save(row).Error; err != nil {
			return nil, err
		}
		return row, nil
	}
	existing, err := GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return update(existing)
	}
	row := UserProfile{HeightIn: heightIn, Age: age, Sex: sex, ActivityLevel: activityLevel}
	if err := database.CreateInSavepoint(db, &row); err != nil {
		// Lost a race with a concurrent first save: update the winner's row.
		if winner, gerr := GetProfile(ctx); gerr == nil && winner != nil {
			return update(winner)
		}
		return nil, err
	}
	return &row, nil
}
