package services

import (
	"be-simpletracker/internal/utils"
	"be-simpletracker/internal/workout/models"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestUpsertCardioForWorkoutLog_usesPlannedTypeWhenTypeEmpty(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.Exercise{},
		&models.WorkoutPlan{},
		&models.WorkoutLog{},
		&models.LoggedExercise{},
		&models.LoggedSet{},
		&models.Cardio{},
	); err != nil {
		t.Fatal(err)
	}
	today := utils.ZerodTime(0)
	dow := int(today.Weekday())
	plan := models.WorkoutPlan{Name: "Test", DayOfWeek: &dow, PlannedCardioType: "Bike"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today, WorkoutPlanID: &plan.ID}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	c, err := UpsertCardioForWorkoutLog(db, 0, 25, "", "S3E4 of Breaking Bad")
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != "Bike" || c.Minutes != 25 || c.Notes != "S3E4 of Breaking Bad" {
		t.Fatalf("got %+v", c)
	}
}
