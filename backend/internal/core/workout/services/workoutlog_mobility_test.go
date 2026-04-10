package services

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/utils"
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestUpsertMobilityPre_persistsChecked(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
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
	if err := db.AutoMigrate(&models.WorkoutPlanExercise{}); err != nil {
		t.Fatal(err)
	}
	today := utils.ZerodTime(0)
	plan := models.WorkoutPlan{
		Name:             "MobilityTestPlan",
		PreMobilityItems: []string{"A", "B", "C"},
	}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today, WorkoutPlanID: &plan.ID}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewWorkoutLogService(db)
	view, err := svc.UpsertMobilityPre(context.Background(), 0, []string{"A", "C"})
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Checked) != 2 || view.Checked[0] != "A" || view.Checked[1] != "C" {
		t.Fatalf("unexpected view %+v", view)
	}
	var reloaded models.WorkoutLog
	if err := db.First(&reloaded, wl.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(reloaded.PreMobilityChecked) != 2 {
		t.Fatalf("db %+v", reloaded.PreMobilityChecked)
	}
}
