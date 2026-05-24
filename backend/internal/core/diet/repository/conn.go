package dietrepo

import (
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

func conn() *gorm.DB {
	return database.GetDB()
}
