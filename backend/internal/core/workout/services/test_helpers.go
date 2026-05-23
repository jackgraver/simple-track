package services

import (
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

func useTestDB(db *gorm.DB) {
	database.SetDB(db)
}
