package steps

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const defaultListLimit = 365

func UpsertSteps(db *gorm.DB, date time.Time, stepsVal int) (*StepLog, error) {
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

func ListSteps(db *gorm.DB, limit int) ([]StepLog, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	var rows []StepLog
	err := db.Order("date DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
