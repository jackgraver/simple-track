package services_test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/utils"
	"strings"
	"testing"
)

func TestAddExerciseToPlan_RemoveExerciseFromPlan_Reorder(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ex1, err := services.CreateExercise("A", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	ex2, err := services.CreateExercise("B", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	plan := models.WorkoutPlan{Name: "Full Body"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	if err := services.AddExerciseToPlan(plan.ID, ex1.ID); err != nil {
		t.Fatal(err)
	}
	if err := services.AddExerciseToPlan(plan.ID, ex2.ID); err != nil {
		t.Fatal(err)
	}
	if err := services.AddExerciseToPlan(plan.ID, ex1.ID); err == nil || !strings.Contains(err.Error(), "already in plan") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	if err := services.ReorderPlanExercises(plan.ID, []uint{ex2.ID, ex1.ID}); err != nil {
		t.Fatal(err)
	}
	loaded, err := services.LoadPlanWithOrderedExercises(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Exercises) != 2 || loaded.Exercises[0].Name != "B" || loaded.Exercises[1].Name != "A" {
		t.Fatalf("order %+v", loaded.Exercises)
	}
	if err := services.RemoveExerciseFromPlan(plan.ID, ex2.ID); err != nil {
		t.Fatal(err)
	}
	loaded, err = services.LoadPlanWithOrderedExercises(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Exercises) != 1 || loaded.Exercises[0].Name != "A" {
		t.Fatalf("after remove %+v", loaded.Exercises)
	}
}

func TestAssignPlanToDay_andUnassign(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := models.WorkoutPlan{Name: "Upper"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	today := utils.ZerodTime(0)
	dow := int(today.Weekday())
	assigned, err := services.AssignPlanToDay(plan.ID, dow)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.DayOfWeek == nil || *assigned.DayOfWeek != dow {
		t.Fatalf("got %+v", assigned.DayOfWeek)
	}
	byDay, err := services.GetPlanByDay(dow)
	if err != nil || byDay == nil || byDay.ID != plan.ID {
		t.Fatalf("GetPlanByDay got %+v err=%v", byDay, err)
	}
	unassigned, err := services.UnassignPlanFromDay(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if unassigned.DayOfWeek != nil {
		t.Fatalf("expected nil day, got %+v", unassigned.DayOfWeek)
	}
}

func TestAssignPlanToDay_rejectsInvalidDay(t *testing.T) {
	testutil.SetupTestDB(t)
	_, err := services.AssignPlanToDay(1, 9)
	if err == nil || !strings.Contains(err.Error(), "day_of_week") {
		t.Fatalf("expected day_of_week error, got %v", err)
	}
}

func TestSetPlannedCardio(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := models.WorkoutPlan{Name: "Cardio Day"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	updated, err := services.SetPlannedCardio(plan.ID, "  Row  ", 30)
	if err != nil {
		t.Fatal(err)
	}
	if updated.PlannedCardioType != "Row" || updated.PlannedCardioMinutes != 30 {
		t.Fatalf("got %q", updated.PlannedCardioType)
	}
}

func TestGetAllWorkoutPlans(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if err := db.Create(&models.WorkoutPlan{Name: "One"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.WorkoutPlan{Name: "Two"}).Error; err != nil {
		t.Fatal(err)
	}
	plans, err := services.GetAllWorkoutPlans()
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 2 {
		t.Fatalf("got %d plans", len(plans))
	}
}
