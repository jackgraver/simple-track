package db

import (
	"be-simpletracker/mealplanner"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectToDB initializes the database connection and runs auto-migration
func ConnectToDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=pass123 dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}

	mealplanner.MigrateMealPlanDatabase(db)

	return db
}