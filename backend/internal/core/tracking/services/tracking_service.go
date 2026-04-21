package services

import (
	"errors"
	"time"

	"be-simpletracker/internal/core/tracking/models"
	"be-simpletracker/internal/utils"

	"gorm.io/gorm"
)

const defaultListLimit = 365

// ParseDateString parses YYYY-MM-DD in local time, or returns today at midnight if empty.
func ParseDateString(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return utils.ZerodTime(0), nil
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
}

func UpsertBodyWeight(db *gorm.DB, date time.Time, weightLbs float64) (*models.BodyWeightLog, error) {
	if weightLbs <= 0 {
		return nil, errors.New("weight_lbs must be positive")
	}
	var row models.BodyWeightLog
	err := db.Where("date = ?", date).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.BodyWeightLog{Date: date, WeightLbs: weightLbs}
		if err := db.Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	row.WeightLbs = weightLbs
	if err := db.Save(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func ListBodyWeights(db *gorm.DB, limit int) ([]models.BodyWeightLog, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	var rows []models.BodyWeightLog
	err := db.Order("date DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func UpsertSteps(db *gorm.DB, date time.Time, steps int) (*models.StepLog, error) {
	if steps < 0 {
		return nil, errors.New("steps must be non-negative")
	}
	var row models.StepLog
	err := db.Where("date = ?", date).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.StepLog{Date: date, Steps: steps}
		if err := db.Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	if err != nil {
		return nil, err
	}
	row.Steps = steps
	if err := db.Save(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func ListSteps(db *gorm.DB, limit int) ([]models.StepLog, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	var rows []models.StepLog
	err := db.Order("date DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
