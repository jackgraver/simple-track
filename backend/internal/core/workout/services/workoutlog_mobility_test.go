package services_test

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"strings"
	"testing"
)

func TestUpsertMobilityPre_persistsChecked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := testutil.Ctx(db)
	today := clock.ZerodTime(0)
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
	view, err := services.UpsertMobilityPre(ctx, 0, []string{"A", "C", "invalid"})
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Checked) != 2 || view.Checked[0] != "A" || view.Checked[1] != "C" {
		t.Fatalf("unexpected view %+v", view)
	}
	if view.Title != "Dynamic warmup" {
		t.Fatalf("title got %q", view.Title)
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
	db := testutil.SetupTestDB(t)
	ctx := testutil.Ctx(db)
	today := clock.ZerodTime(0)
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
	view, err := services.UpsertMobilityPost(ctx, 0, []string{"Y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Checked) != 1 || view.Checked[0] != "Y" {
		t.Fatalf("unexpected view %+v", view)
	}
	if view.Title != "Static stretching" {
		t.Fatalf("title got %q", view.Title)
	}
}

func TestUpsertMobilityPre_errorsWhenNothingPlanned(t *testing.T) {
	ctx := testutil.Ctx(testutil.SetupTestDB(t))
	_, err := services.UpsertMobilityPre(ctx, 0, []string{"A"})
	if err == nil || !strings.Contains(err.Error(), "no dynamic warmup planned") {
		t.Fatalf("expected no mobility error, got %v", err)
	}
}
