package steps

import (
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

const defaultListLimit = 365

func UpsertSteps(ctx context.Context, date time.Time, stepsVal int) (*StepLog, error) {
	db := database.From(ctx)
	if stepsVal < 0 {
		return nil, errors.New("steps must be non-negative")
	}
	var row StepLog
	err := db.Where("date = ?", date).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = StepLog{Date: date, Steps: stepsVal}
		if err := db.Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	row.Steps = stepsVal
	if err := db.Save(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func ListSteps(ctx context.Context, limit int) ([]StepLog, error) {
	db := database.From(ctx)
	if limit <= 0 {
		limit = defaultListLimit
	}
	var rows []StepLog
	err := db.Order("date DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
