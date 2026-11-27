package services_test

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/database"
	"context"
	"slices"
	"strings"
	"testing"
)

func TestAddExerciseToPlan_RemoveExerciseFromPlan_Reorder(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	ex1, err := services.CreateExercise(ctx, "A", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	ex2, err := services.CreateExercise(ctx, "B", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	plan := models.WorkoutPlan{Name: "Full Body"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	if err := services.AddExerciseToPlan(ctx, plan.ID, ex1.ID); err != nil {
		t.Fatal(err)
	}
	if err := services.AddExerciseToPlan(ctx, plan.ID, ex2.ID); err != nil {
		t.Fatal(err)
	}
	if err := services.AddExerciseToPlan(ctx, plan.ID, ex1.ID); err == nil || !strings.Contains(err.Error(), "already in plan") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	if err := services.ReorderPlanExercises(ctx, plan.ID, []uint{ex2.ID, ex1.ID}); err != nil {
		t.Fatal(err)
	}
	loaded, err := services.LoadPlanWithOrderedExercises(ctx, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Exercises) != 2 || loaded.Exercises[0].Name != "B" || loaded.Exercises[1].Name != "A" {
		t.Fatalf("order %+v", loaded.Exercises)
	}
	if err := services.RemoveExerciseFromPlan(ctx, plan.ID, ex2.ID); err != nil {
		t.Fatal(err)
	}
	loaded, err = services.LoadPlanWithOrderedExercises(ctx, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Exercises) != 1 || loaded.Exercises[0].Name != "A" {
		t.Fatalf("after remove %+v", loaded.Exercises)
	}
}

func TestAssignPlanToDay_andUnassign(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	plan := models.WorkoutPlan{Name: "Upper"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	today := clock.ZerodTime(0)
	dow := int(today.Weekday())
	assigned, err := services.AssignPlanToDay(ctx, plan.ID, dow)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.DayOfWeek == nil || *assigned.DayOfWeek != dow {
		t.Fatalf("got %+v", assigned.DayOfWeek)
	}
	byDay, err := services.GetPlanByDay(ctx, dow)
	if err != nil || byDay == nil || byDay.ID != plan.ID {
		t.Fatalf("GetPlanByDay got %+v err=%v", byDay, err)
	}
	unassigned, err := services.UnassignPlanFromDay(ctx, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if unassigned.DayOfWeek != nil {
		t.Fatalf("expected nil day, got %+v", unassigned.DayOfWeek)
	}
}

func TestAssignPlanToDay_rejectsInvalidDay(t *testing.T) {
	ctx := database.WithDB(context.Background(), testutil.SetupTestDB(t))
	_, err := services.AssignPlanToDay(ctx, 1, 9)
	if err == nil || !strings.Contains(err.Error(), "day_of_week") {
		t.Fatalf("expected day_of_week error, got %v", err)
	}
}

func TestSetPlannedCardio(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	plan := models.WorkoutPlan{Name: "Cardio Day"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	updated, err := services.SetPlannedCardio(ctx, plan.ID, "  Row  ", 30)
	if err != nil {
		t.Fatal(err)
	}
	if updated.PlannedCardioType != "Row" || updated.PlannedCardioMinutes != 30 {
		t.Fatalf("got %q", updated.PlannedCardioType)
	}
}

func TestSetPlannedMobility(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	plan := models.WorkoutPlan{Name: "Recovery Day"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	updated, err := services.SetPlannedMobility(ctx,
		plan.ID,
		[]string{" Leg swings ", "", "leg swings", "Arm circles"},
		[]string{"Hamstring stretch", " Quad stretch "},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(updated.PreMobilityItems, []string{"Leg swings", "Arm circles"}) {
		t.Fatalf("dynamic warmup got %#v", updated.PreMobilityItems)
	}
	if !slices.Equal(updated.PostMobilityItems, []string{"Hamstring stretch", "Quad stretch"}) {
		t.Fatalf("static stretching got %#v", updated.PostMobilityItems)
	}
	cleared, err := services.SetPlannedMobility(ctx, plan.ID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.PreMobilityItems) != 0 || len(cleared.PostMobilityItems) != 0 {
		t.Fatalf("mobility items not cleared: %#v %#v", cleared.PreMobilityItems, cleared.PostMobilityItems)
	}
}

func TestGetAllWorkoutPlans(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	if err := db.Create(&models.WorkoutPlan{Name: "One"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.WorkoutPlan{Name: "Two"}).Error; err != nil {
		t.Fatal(err)
	}
	plans, err := services.GetAllWorkoutPlans(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 2 {
		t.Fatalf("got %d plans", len(plans))
	}
}
