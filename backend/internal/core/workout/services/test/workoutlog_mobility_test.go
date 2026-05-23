package test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils"
	"context"
	"strings"
	"testing"
)

func TestUpsertMobilityPre_persistsChecked(t *testing.T) {
	db := setupTestDB(t)
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
	view, err := services.UpsertMobilityPre(context.Background(), 0, []string{"A", "C", "invalid"})
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

func TestUpsertMobilityPost_persistsChecked(t *testing.T) {
	db := setupTestDB(t)
	today := utils.ZerodTime(0)
	plan := models.WorkoutPlan{
		Name:              "MobilityTestPlan",
		PostMobilityItems: []string{"X", "Y"},
	}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today, WorkoutPlanID: &plan.ID}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	view, err := services.UpsertMobilityPost(context.Background(), 0, []string{"Y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Checked) != 1 || view.Checked[0] != "Y" {
		t.Fatalf("unexpected view %+v", view)
	}
}

func TestUpsertMobilityPre_errorsWhenNothingPlanned(t *testing.T) {
	setupTestDB(t)
	_, err := services.UpsertMobilityPre(context.Background(), 0, []string{"A"})
	if err == nil || !strings.Contains(err.Error(), "no pre-workout mobility planned") {
		t.Fatalf("expected no mobility error, got %v", err)
	}
}
