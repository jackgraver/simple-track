package test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils"
	"context"
	"errors"
	"testing"
)

func TestGetWorkoutActivity_year_includesDayWithSet(t *testing.T) {
	db := setupTestDB(t)
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
	res, err := services.GetWorkoutActivity(context.Background(), "year", 52)
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
	db := setupTestDB(t)
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
	res, err := services.GetWorkoutActivity(context.Background(), "rolling", 52)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range res.ActiveDates {
		if d == old.Format("2006-01-02") {
			t.Fatalf("old day should be outside window: %+v", res.ActiveDates)
		}
	}
}

func TestGetWorkoutActivity_invalidMode(t *testing.T) {
	setupTestDB(t)
	_, err := services.GetWorkoutActivity(context.Background(), "nope", 52)
	if !errors.Is(err, services.ErrInvalidActivityMode) {
		t.Fatalf("expected ErrInvalidActivityMode, got %v", err)
	}
}
