package weight

import (
	"be-simpletracker/internal/core/tracking/common"
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultListLimit = 365

// UpsertBodyWeight sets the current user's weight for date: one row per (user, date).
func UpsertBodyWeight(ctx context.Context, date time.Time, weightLbs float64) (*BodyWeightLog, error) {
	if weightLbs <= 0 {
		return nil, common.Invalid("weight_lbs must be positive")
	}
	db := database.From(ctx)
	find := func() (*BodyWeightLog, error) {
		var row BodyWeightLog
		if err := db.Scopes(database.Owned(ctx)).Where("date = ?", date).First(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	update := func(row *BodyWeightLog) (*BodyWeightLog, error) {
		row.WeightLbs = weightLbs
		if err := db.Omit(clause.Associations).Save(row).Error; err != nil {
			return nil, err
		}
		return row, nil
	}
	row, err := find()
	switch {
	case err == nil:
		return update(row)
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}
	created := BodyWeightLog{Date: date, WeightLbs: weightLbs}
	if err := database.CreateInSavepoint(db, &created); err != nil {
		// Lost a race with a concurrent first weigh-in for the same day: update the winner's row.
		if row, ferr := find(); ferr == nil {
			return update(row)
		}
		return nil, err
	}
	return &created, nil
}

// ListBodyWeights returns the current user's logs, newest first.
func ListBodyWeights(ctx context.Context, limit int) ([]BodyWeightLog, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	var rows []BodyWeightLog
	err := database.From(ctx).Scopes(database.Owned(ctx)).Order("date DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
