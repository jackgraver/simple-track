package services_test

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/database"
	"context"
	"testing"
)

func TestRenameWorkoutPlan(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	var program models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&program).Error; err != nil {
		t.Fatal(err)
	}
	plan := models.WorkoutPlan{Name: "Push", WorkoutProgramID: &program.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	renamed, err := services.RenameWorkoutPlan(ctx, plan.ID, "  Push A ")
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Name != "Push A" {
		t.Fatalf("got %q", renamed.Name)
	}
	if _, err := services.RenameWorkoutPlan(ctx, plan.ID, " "); err == nil {
		t.Fatal("expected error for blank name")
	}
	if _, err := services.RenameWorkoutPlan(ctx, 9999, "X"); err == nil {
		t.Fatal("expected error for missing routine")
	}
}

func TestDeleteAndRestoreWorkoutPlanKeepsExercisesAndFreeDays(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
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
	if err := services.AddExerciseToPlan(ctx, push.ID, bench.ID); err != nil {
		t.Fatal(err)
	}
	for _, d := range []int{1, 4} {
		if _, err := services.AssignPlanToDay(ctx, push.ID, d); err != nil {
			t.Fatal(err)
		}
	}

	if err := services.DeleteWorkoutPlan(ctx, push.ID); err != nil {
		t.Fatal(err)
	}
	plans, err := services.GetAllWorkoutPlans(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range plans {
		if p.ID == push.ID {
			t.Fatal("deleted routine still listed")
		}
	}
	// Monday goes to Pull while Push is deleted.
	if _, err := services.AssignPlanToDay(ctx, pull.ID, 1); err != nil {
		t.Fatal(err)
	}

	restored, err := services.RestoreWorkoutPlan(ctx, push.ID)
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
	if _, err := services.RestoreWorkoutPlan(ctx, push.ID); err == nil {
		t.Fatal("restoring a live routine should fail")
	}
}

func TestLogKeepsRoutineAfterDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	plan := models.WorkoutPlan{Name: "Legs"}
	squat := models.Exercise{Name: "Squat"}
	for _, v := range []any{&plan, &squat} {
		if err := db.Create(v).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := services.AddExerciseToPlan(ctx, plan.ID, squat.ID); err != nil {
		t.Fatal(err)
	}
	past := models.WorkoutLog{Date: clock.ZerodTime(3), WorkoutPlanID: &plan.ID}
	if err := db.Create(&past).Error; err != nil {
		t.Fatal(err)
	}
	if err := services.DeleteWorkoutPlan(ctx, plan.ID); err != nil {
		t.Fatal(err)
	}
	got, err := services.GetOrCreateToday(ctx, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkoutPlan == nil || got.WorkoutPlan.Name != "Legs" {
		t.Fatalf("log lost its deleted routine: %+v", got.WorkoutPlan)
	}
	if len(got.WorkoutPlan.Exercises) != 1 || got.WorkoutPlan.Exercises[0].ID != squat.ID {
		t.Fatalf("exercises missing: %+v", got.WorkoutPlan.Exercises)
	}
}
