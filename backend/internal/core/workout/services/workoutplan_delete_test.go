package services_test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"testing"
)

func TestRenameWorkoutPlan(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var program models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&program).Error; err != nil {
		t.Fatal(err)
	}
	plan := models.WorkoutPlan{Name: "Push", WorkoutProgramID: &program.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	renamed, err := services.RenameWorkoutPlan(plan.ID, "  Push A ")
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Name != "Push A" {
		t.Fatalf("got %q", renamed.Name)
	}
	if _, err := services.RenameWorkoutPlan(plan.ID, " "); err == nil {
		t.Fatal("expected error for blank name")
	}
	if _, err := services.RenameWorkoutPlan(9999, "X"); err == nil {
		t.Fatal("expected error for missing routine")
	}
}

func TestDeleteAndRestoreWorkoutPlanKeepsExercisesAndFreeDays(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var program models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&program).Error; err != nil {
		t.Fatal(err)
	}
	push := models.WorkoutPlan{Name: "Push", WorkoutProgramID: &program.ID}
	pull := models.WorkoutPlan{Name: "Pull", WorkoutProgramID: &program.ID}
	bench := models.Exercise{Name: "Bench"}
	for _, v := range []any{&push, &pull, &bench} {
		if err := db.Create(v).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := services.AddExerciseToPlan(push.ID, bench.ID); err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{1, 4} {
		if _, err := services.AssignPlanToDay(push.ID, d); err != nil {
			t.Fatal(err)
		}
	}

	if err := services.DeleteWorkoutPlan(push.ID); err != nil {
		t.Fatal(err)
	}
	plans, err := services.GetAllWorkoutPlans()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range plans {
		if p.ID == push.ID {
			t.Fatal("deleted routine still listed")
		}
	}
	// Monday goes to Pull while Push is deleted.
	if _, err := services.AssignPlanToDay(pull.ID, 1); err != nil {
		t.Fatal(err)
	}

	restored, err := services.RestoreWorkoutPlan(push.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.Exercises) != 1 || restored.Exercises[0].ID != bench.ID {
		t.Fatalf("exercises not kept: %+v", restored.Exercises)
	}
	if len(restored.AssignedDays) != 1 || restored.AssignedDays[0] != 4 {
		t.Fatalf("expected only Thursday back, got %v", restored.AssignedDays)
	}
	if restored.DayOfWeek == nil || *restored.DayOfWeek != 4 {
		t.Fatalf("legacy day not refreshed: %v", restored.DayOfWeek)
	}
	if _, err := services.RestoreWorkoutPlan(push.ID); err == nil {
		t.Fatal("restoring a live routine should fail")
	}
}
