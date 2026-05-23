package test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils"
	"context"
	"strings"
	"testing"
)

func TestUpsertCardioForWorkoutLog_usesPlannedTypeWhenTypeEmpty(t *testing.T) {
	db := setupTestDB(t)
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
	c, err := services.UpsertCardioForWorkoutLog(context.Background(), 0, 25, "", "S3E4 of Breaking Bad")
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != "Bike" || c.Minutes != 25 || c.Notes != "S3E4 of Breaking Bad" {
		t.Fatalf("got %+v", c)
	}
}

func TestUpsertCardio_updatesExistingRow(t *testing.T) {
	db := setupTestDB(t)
	today := utils.ZerodTime(0)
	wl := models.WorkoutLog{Date: today}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	existing := models.Cardio{WorkoutLogID: wl.ID, Minutes: 10, Type: "Run", Notes: "easy"}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	c, err := services.UpsertCardio(context.Background(), 0, 30, "Bike", "hard")
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != existing.ID || c.Minutes != 30 || c.Type != "Bike" || c.Notes != "hard" {
		t.Fatalf("got %+v", c)
	}
}

func TestUpsertCardio_requiresTypeWhenNoPlan(t *testing.T) {
	setupTestDB(t)
	_, err := services.UpsertCardio(context.Background(), 0, 20, "", "")
	if err == nil || !strings.Contains(err.Error(), "cardio type is required") {
		t.Fatalf("expected type required error, got %v", err)
	}
}
