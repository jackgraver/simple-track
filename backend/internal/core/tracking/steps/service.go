package steps

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

// UpsertSteps sets the current user's step count for date: one row per (user, date).
func UpsertSteps(ctx context.Context, date time.Time, stepsVal int) (*StepLog, error) {
	if stepsVal < 0 {
		return nil, common.Invalid("steps must be non-negative")
	}
	db := database.From(ctx)
	find := func() (*StepLog, error) {
		var row StepLog
		if err := db.Scopes(database.Owned(ctx)).Where("date = ?", date).First(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	update := func(row *StepLog) (*StepLog, error) {
		row.Steps = stepsVal
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
	created := StepLog{Date: date, Steps: stepsVal}
	if err := database.CreateInSavepoint(db, &created); err != nil {
		// Lost a race with a concurrent first entry for the same day: update the winner's row.
		if row, ferr := find(); ferr == nil {
			return update(row)
		}
		return nil, err
	}
	return &created, nil
}

// ListSteps returns the current user's logs, newest first.
func ListSteps(ctx context.Context, limit int) ([]StepLog, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	var rows []StepLog
	err := database.From(ctx).Scopes(database.Owned(ctx)).Order("date DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
