package services_test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"context"
	"testing"
	"time"
)

func TestActiveWorkoutProgramResolvesItsOwnWeek(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var current models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&current).Error; err != nil {
		t.Fatal(err)
	}
	other := models.WorkoutProgram{Name: "Upper Lower"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	day := 1
	currentPlan := models.WorkoutPlan{Name: "Current Push", WorkoutProgramID: &current.ID, DayOfWeek: &day}
	otherPlan := models.WorkoutPlan{Name: "Other Push", WorkoutProgramID: &other.ID, DayOfWeek: &day}
	if err := db.Create(&currentPlan).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&otherPlan).Error; err != nil {
		t.Fatal(err)
	}
	plan, err := services.GetPlanByDay(day)
	if err != nil {
		t.Fatal(err)
	}
	if plan == nil || plan.ID != currentPlan.ID {
		t.Fatalf("expected current program plan, got %+v", plan)
	}
}

func TestActivatingWorkoutProgramPreservesExistingDayLog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var current models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&current).Error; err != nil {
		t.Fatal(err)
	}
	next := models.WorkoutProgram{Name: "Next", IsActive: false}
	if err := db.Create(&next).Error; err != nil {
		t.Fatal(err)
	}
	day := int(time.Now().Weekday())
	plan := models.WorkoutPlan{Name: "Current", WorkoutProgramID: &current.ID, DayOfWeek: &day}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := services.ActivateWorkoutProgram(context.Background(), next.ID); err != nil {
		t.Fatal(err)
	}
	var currentLog models.WorkoutLog
	if err := db.Order("id DESC").First(&currentLog).Error; err != nil {
		t.Fatal(err)
	}
	if currentLog.WorkoutPlanID == nil || *currentLog.WorkoutPlanID != plan.ID {
		t.Fatalf("expected today's log to preserve plan %d, got %+v", plan.ID, currentLog.WorkoutPlanID)
	}
	var active models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&active).Error; err != nil {
		t.Fatal(err)
	}
	if active.ID != next.ID {
		t.Fatalf("expected program %d active, got %d", next.ID, active.ID)
	}
}

func TestWorkoutPlanCanBeAssignedToMultipleWeekdays(t *testing.T) {
	db := testutil.SetupTestDB(t)
	var program models.WorkoutProgram
	if err := db.Where("is_active = ?", true).First(&program).Error; err != nil {
		t.Fatal(err)
	}
	plan := models.WorkoutPlan{Name: "Push", WorkoutProgramID: &program.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := services.AssignPlanToDay(plan.ID, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := services.AssignPlanToDay(plan.ID, 4); err != nil {
		t.Fatal(err)
	}
	loaded, err := services.LoadPlanWithOrderedExercises(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.AssignedDays) != 2 || loaded.AssignedDays[0] != 1 || loaded.AssignedDays[1] != 4 {
		t.Fatalf("expected Monday and Thursday, got %v", loaded.AssignedDays)
	}
}

func TestCreateWorkoutPlanWithoutDayCreatesUnassignedRoutine(t *testing.T) {
	testutil.SetupTestDB(t)
	program, err := services.CreateWorkoutProgram("New Split")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := services.CreateWorkoutPlan(program.ID, "Push", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.AssignedDays) != 0 || plan.DayOfWeek != nil {
		t.Fatalf("expected an unassigned routine, got days=%v legacy_day=%v", plan.AssignedDays, plan.DayOfWeek)
	}
}
