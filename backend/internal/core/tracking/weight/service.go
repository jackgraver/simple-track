package weight

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const defaultListLimit = 365

func UpsertBodyWeight(db *gorm.DB, date time.Time, weightLbs float64) (*BodyWeightLog, error) {
	if weightLbs <= 0 {
		return nil, errors.New("weight_lbs must be positive")
	}
	var row BodyWeightLog
	err := db.Where("date = ?", date).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = BodyWeightLog{Date: date, WeightLbs: weightLbs}
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

func ListBodyWeights(db *gorm.DB, limit int) ([]BodyWeightLog, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	var rows []BodyWeightLog
	err := db.Order("date DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
