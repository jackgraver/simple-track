package testutil

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/database"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.Exercise{},
		&models.WorkoutProgram{},
		&models.WorkoutPlan{},
		&models.WorkoutPlanDay{},
		&models.WorkoutPlanExercise{},
		&models.WorkoutLog{},
		&models.LoggedExercise{},
		&models.LoggedSet{},
		&models.Cardio{},
	); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.WorkoutProgram{Name: "Test Program", IsActive: true}).Error; err != nil {
		t.Fatal(err)
	}
	database.SetDB(db)
	return db
}
