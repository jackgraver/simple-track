package db

import (
	"be-simpletracker/mealplanner"
	"be-simpletracker/workout"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Initializes a new database connection to a PostgreSQL database
func ConnectToPostgres() *gorm.DB {
	dsn := "host=localhost user=postgres password=pass123 dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}

	mealplanner.MigrateMealPlanDatabase(db)
	workout.MigrateWorkoutDatabase(db)

	return db
}

// Initializes a new database connection to a SQLite database
func ConnectToSqlite() *gorm.DB {
    db, err := gorm.Open(sqlite.Open("st.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database: " + err.Error())
    }

    mealplanner.MigrateMealPlanDatabase(db)
    workout.MigrateWorkoutDatabase(db)

    return db
}