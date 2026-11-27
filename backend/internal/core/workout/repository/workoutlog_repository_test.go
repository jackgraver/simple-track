package workoutrepo_test

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/database"
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestUpdateWorkoutPlanID_clearsPlan(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	plan := models.WorkoutPlan{Name: "Temp"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{WorkoutPlanID: &plan.ID}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	if err := workoutrepo.UpdateWorkoutPlanID(ctx, wl.ID, nil); err != nil {
		t.Fatal(err)
	}
	var reloaded models.WorkoutLog
	if err := db.First(&reloaded, wl.ID).Error; err != nil {
		t.Fatal(err)
	}
	if reloaded.WorkoutPlanID != nil {
		t.Fatalf("expected nil plan id, got %+v", reloaded.WorkoutPlanID)
	}
}

func TestUpdatePostMobilityChecked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	wl := models.WorkoutLog{}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	checked := []string{"hip", "shoulder"}
	if err := workoutrepo.UpdatePostMobilityChecked(ctx, wl.ID, checked); err != nil {
		t.Fatal(err)
	}
	var reloaded models.WorkoutLog
	if err := db.First(&reloaded, wl.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(reloaded.PostMobilityChecked) != 2 {
		t.Fatalf("got %+v", reloaded.PostMobilityChecked)
	}
}

func TestUpdatePreMobilityChecked_notFound(t *testing.T) {
	ctx := database.WithDB(context.Background(), testutil.SetupTestDB(t))
	err := workoutrepo.UpdatePreMobilityChecked(ctx, 9999, []string{"a"})
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
