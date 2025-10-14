package services

import (
	"be-simpletracker/db/models"
	"time"

	"gorm.io/gorm"
)

//TODO think of transaction handling architecture

func ObjectRange[T models.Preloadable](db *gorm.DB, start time.Time, end time.Time) ([]T, error) {
	var objects []T
	var t T

	tx := db
	for _, p := range t.Preloads() {
		tx = tx.Preload(p)
	}

	if err := tx.
		Where("date BETWEEN ? AND ?", start, end).
		Order("date").
		Find(&objects).Error; err != nil {
		return nil, err
	}
	return objects, nil
}