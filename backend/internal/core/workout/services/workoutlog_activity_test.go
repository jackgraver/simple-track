package services

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/utils"
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGetWorkoutActivity_year_includesDayWithSet(t *testing.T) {
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
	today := utils.ZerodTime(0)
	ex := models.Exercise{Name: "Bench"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	le := models.LoggedExercise{WorkoutLogID: wl.ID, ExerciseID: ex.ID}
	if err := db.Create(&le).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LoggedSet{LoggedExerciseID: le.ID, Reps: 5, Weight: 100}).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewWorkoutLogService(db)
	res, err := svc.GetWorkoutActivity(context.Background(), "year", 52)
	if err != nil {
		t.Fatal(err)
	}
	want := today.Format("2006-01-02")
	found := false
	for _, d := range res.ActiveDates {
		if d == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %s in %+v", want, res.ActiveDates)
	}
	if res.Mode != "year" {
		t.Fatalf("mode %q", res.Mode)
	}
}

func TestGetWorkoutActivity_rolling_excludesOldDayOutsideWindow(t *testing.T) {
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
	ex := models.Exercise{Name: "Squat"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	old := utils.ZerodTime(400)
	wlOld := models.WorkoutLog{Date: old}
	if err := db.Create(&wlOld).Error; err != nil {
		t.Fatal(err)
	}
	leOld := models.LoggedExercise{WorkoutLogID: wlOld.ID, ExerciseID: ex.ID}
	if err := db.Create(&leOld).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LoggedSet{LoggedExerciseID: leOld.ID, Reps: 3, Weight: 50}).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewWorkoutLogService(db)
	res, err := svc.GetWorkoutActivity(context.Background(), "rolling", 52)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range res.ActiveDates {
		if d == old.Format("2006-01-02") {
			t.Fatalf("old day should be outside 30d window: %+v", res.ActiveDates)
		}
	}
}

func TestGetWorkoutActivity_invalidMode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.WorkoutLog{}); err != nil {
		t.Fatal(err)
	}
	svc := NewWorkoutLogService(db)
	_, err = svc.GetWorkoutActivity(context.Background(), "nope", 52)
	if !errors.Is(err, ErrInvalidActivityMode) {
		t.Fatalf("expected ErrInvalidActivityMode, got %v", err)
	}
}
