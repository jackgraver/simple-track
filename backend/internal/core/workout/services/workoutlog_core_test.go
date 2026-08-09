package services_test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/utils"
	"context"
	"strings"
	"testing"
)

func TestGetOrCreateToday_returnsExistingLog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	wl := models.WorkoutLog{Date: today}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	got, err := services.GetOrCreateToday(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != wl.ID {
		t.Fatalf("got id %d want %d", got.ID, wl.ID)
	}
}

func TestGetOrCreateToday_createsLogWithPlanForDay(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	dow := int(today.Weekday())
	plan := models.WorkoutPlan{Name: "Push", DayOfWeek: &dow}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	got, err := services.GetOrCreateToday(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkoutPlanID == nil || *got.WorkoutPlanID != plan.ID {
		t.Fatalf("expected plan id %d, got %+v", plan.ID, got.WorkoutPlanID)
	}
}

func TestSwitchPlan_updatesPlanAndReturnsView(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	dow := int(today.Weekday())
	plan := models.WorkoutPlan{Name: "Legs", DayOfWeek: &dow, PlannedCardioType: "Run", PlannedCardioMinutes: 30}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	ex := models.Exercise{Name: "Squat"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.WorkoutPlanExercise{
		WorkoutPlanID: plan.ID,
		ExerciseID:    ex.ID,
		DisplayOrder:  0,
	}).Error; err != nil {
		t.Fatal(err)
	}
	planID := plan.ID
	res, err := services.SwitchPlan(context.Background(), 0, &planID)
	if err != nil {
		t.Fatal(err)
	}
	if res.Day.WorkoutPlanID == nil || *res.Day.WorkoutPlanID != plan.ID {
		t.Fatalf("day plan %+v", res.Day.WorkoutPlanID)
	}
	if res.PlannedCardio == nil {
		t.Fatal("expected planned cardio")
	}
	cardio, ok := res.PlannedCardio.(map[string]any)
	if !ok || cardio["minutes"] != 30 {
		t.Fatalf("planned cardio %+v", res.PlannedCardio)
	}
	if len(res.PlannedExercises) != 1 || res.PlannedExercises[0].Planned.Name != "Squat" {
		t.Fatalf("planned exercises %+v", res.PlannedExercises)
	}
}

func TestSwitchPlan_rejectsMissingPlan(t *testing.T) {
	testutil.SetupTestDB(t)
	missing := uint(9999)
	_, err := services.SwitchPlan(context.Background(), 0, &missing)
	if err == nil || !strings.Contains(err.Error(), "workout plan not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetPreviousWorkoutView_includesPreviousLog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	yesterday := utils.ZerodTime(1)
	ex := models.Exercise{Name: "Deadlift"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	plan := models.WorkoutPlan{Name: "Pull"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.WorkoutPlanExercise{
		WorkoutPlanID: plan.ID,
		ExerciseID:    ex.ID,
		DisplayOrder:  0,
	}).Error; err != nil {
		t.Fatal(err)
	}
	wlPrev := models.WorkoutLog{Date: yesterday, WorkoutPlanID: &plan.ID}
	if err := db.Create(&wlPrev).Error; err != nil {
		t.Fatal(err)
	}
	lePrev := models.LoggedExercise{WorkoutLogID: wlPrev.ID, ExerciseID: ex.ID}
	if err := db.Create(&lePrev).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LoggedSet{LoggedExerciseID: lePrev.ID, Reps: 5, Weight: 140}).Error; err != nil {
		t.Fatal(err)
	}
	wlToday := models.WorkoutLog{Date: today, WorkoutPlanID: &plan.ID}
	if err := db.Create(&wlToday).Error; err != nil {
		t.Fatal(err)
	}
	leToday := models.LoggedExercise{WorkoutLogID: wlToday.ID, ExerciseID: ex.ID}
	if err := db.Create(&leToday).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LoggedSet{LoggedExerciseID: leToday.ID, Reps: 5, Weight: 145}).Error; err != nil {
		t.Fatal(err)
	}
	res, err := services.GetPreviousWorkoutView(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.PlannedExercises) != 1 {
		t.Fatalf("expected 1 group, got %+v", res.PlannedExercises)
	}
	group := res.PlannedExercises[0]
	if group.Logged == nil || group.Previous == nil || group.Max == nil {
		t.Fatalf("expected logged, previous, and max, got %+v", group)
	}
	if len(group.Previous.Sets) != 1 || group.Previous.Sets[0].Weight != 140 {
		t.Fatalf("previous sets %+v", group.Previous.Sets)
	}
	if len(group.Max.Sets) != 1 || group.Max.Sets[0].Weight != 140 {
		t.Fatalf("max sets %+v", group.Max.Sets)
	}
}

func TestGetMonthWorkoutLogs_returnsLogsInRange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	if err := db.Create(&models.WorkoutLog{Date: today}).Error; err != nil {
		t.Fatal(err)
	}
	res, err := services.GetMonthWorkoutLogs(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Days) < 1 {
		t.Fatalf("expected at least one day, got %+v", res.Days)
	}
	if res.Offset != 0 {
		t.Fatalf("offset %d", res.Offset)
	}
}

func TestGetPlanByDay_returnsNilWhenUnassigned(t *testing.T) {
	testutil.SetupTestDB(t)
	plan, err := services.GetPlanByDay(3)
	if err != nil {
		t.Fatal(err)
	}
	if plan != nil {
		t.Fatalf("expected nil plan, got %+v", plan)
	}
}

func TestGetPlanByDay_rejectsInvalidDay(t *testing.T) {
	testutil.SetupTestDB(t)
	_, err := services.GetPlanByDay(7)
	if err == nil || !strings.Contains(err.Error(), "day_of_week") {
		t.Fatalf("expected day_of_week error, got %v", err)
	}
}
