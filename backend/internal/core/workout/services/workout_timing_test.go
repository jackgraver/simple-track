package services_test

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/utils"
	"context"
	"testing"
	"time"
)

var timingDay = time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

// row builds a set saved `at` after 6pm on timingDay.
func row(workout, logged, exercise uint, name string, at time.Duration) workoutrepo.TimedSetRow {
	return workoutrepo.TimedSetRow{
		WorkoutLogID:     workout,
		LogDate:          timingDay,
		LoggedExerciseID: logged,
		ExerciseID:       exercise,
		ExerciseName:     name,
		SetCreatedAt:     timingDay.Add(18*time.Hour + at),
	}
}

func TestComputeWorkoutTimingStats_basic(t *testing.T) {
	m := time.Minute
	rows := []workoutrepo.TimedSetRow{
		// Bench: sets at 0, 3, 6 min. Squat: 10, 12 min.
		row(1, 10, 1, "Bench", 0),
		row(1, 10, 1, "Bench", 3*m),
		row(1, 10, 1, "Bench", 6*m),
		row(1, 11, 2, "Squat", 10*m),
		row(1, 11, 2, "Squat", 12*m),
	}
	got := services.ComputeWorkoutTimingStats(rows)

	if got.Workout.Count != 1 || got.Workout.AvgSeconds != 12*60 {
		t.Fatalf("workout = %+v, want 1 x 720s", got.Workout)
	}
	if got.SetInterval.Count != 3 || got.SetInterval.AvgSeconds != (3+3+2)*60/3.0 {
		t.Fatalf("set interval = %+v", got.SetInterval)
	}
	if got.SetInterval.MedianSeconds != 180 {
		t.Fatalf("set interval median = %v, want 180", got.SetInterval.MedianSeconds)
	}
	if got.Transition.Count != 1 || got.Transition.AvgSeconds != 4*60 {
		t.Fatalf("transition = %+v, want 1 x 240s", got.Transition)
	}
	if len(got.Exercises) != 2 || got.Exercises[0].Name != "Bench" || got.Exercises[1].Name != "Squat" {
		t.Fatalf("exercises = %+v", got.Exercises)
	}
	if got.Exercises[0].Transition.Count != 0 {
		t.Fatalf("first exercise of a session should have no transition: %+v", got.Exercises[0])
	}
	if got.Exercises[1].Transition.AvgSeconds != 240 || got.Exercises[1].SetInterval.AvgSeconds != 120 {
		t.Fatalf("squat = %+v", got.Exercises[1])
	}
}

func TestComputeWorkoutTimingStats_supersetTransitionUsesLatestOtherSet(t *testing.T) {
	m := time.Minute
	rows := []workoutrepo.TimedSetRow{
		row(1, 10, 1, "Curl", 0),
		row(1, 11, 2, "Pushdown", 1*m),
		row(1, 10, 1, "Curl", 2*m),
		row(1, 11, 2, "Pushdown", 3*m),
	}
	got := services.ComputeWorkoutTimingStats(rows)
	if got.Transition.Count != 1 || got.Transition.AvgSeconds != 60 {
		t.Fatalf("transition = %+v, want 1 x 60s", got.Transition)
	}
	if got.SetInterval.Count != 2 || got.SetInterval.AvgSeconds != 120 {
		t.Fatalf("set interval = %+v, want 2 x 120s", got.SetInterval)
	}
}

func TestComputeWorkoutTimingStats_excludesBackfilledAndBulkEntered(t *testing.T) {
	s := time.Second
	backfilled := []workoutrepo.TimedSetRow{
		row(1, 10, 1, "Bench", 72*time.Hour),
		row(1, 10, 1, "Bench", 72*time.Hour+3*time.Minute),
	}
	bulk := []workoutrepo.TimedSetRow{
		row(2, 20, 1, "Bench", 0),
		row(2, 20, 1, "Bench", 4*s),
		row(2, 20, 1, "Bench", 8*s),
		row(2, 21, 2, "Squat", 12*s),
	}
	single := []workoutrepo.TimedSetRow{row(3, 30, 1, "Bench", 0)}
	rows := append(append(backfilled, bulk...), single...)

	got := services.ComputeWorkoutTimingStats(rows)
	if got.Workout.Count != 0 || got.SetInterval.Count != 0 || got.Transition.Count != 0 {
		t.Fatalf("expected nothing counted, got %+v", got)
	}
	if got.ExcludedWorkouts != 3 {
		t.Fatalf("excluded = %d, want 3", got.ExcludedWorkouts)
	}
	if len(got.Exercises) != 0 {
		t.Fatalf("exercises = %+v, want none", got.Exercises)
	}
}

func TestComputeWorkoutTimingStats_dropsOutlierGaps(t *testing.T) {
	m := time.Minute
	rows := []workoutrepo.TimedSetRow{
		row(1, 10, 1, "Bench", 0),
		row(1, 10, 1, "Bench", 3*m),
		row(1, 10, 1, "Bench", 45*m), // left and came back: not a set interval
		row(1, 11, 2, "Squat", 48*m),
	}
	got := services.ComputeWorkoutTimingStats(rows)
	if got.SetInterval.Count != 1 || got.SetInterval.AvgSeconds != 180 {
		t.Fatalf("set interval = %+v, want 1 x 180s", got.SetInterval)
	}
	if got.Workout.AvgSeconds != 48*60 {
		t.Fatalf("workout = %+v", got.Workout)
	}
}

func TestGetWorkoutTimingStats_readsFromDB(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	ex := models.Exercise{Name: "Bench"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	le := models.LoggedExercise{WorkoutLogID: wl.ID, ExerciseID: ex.ID}
	if err := db.Create(&le).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		set := models.LoggedSet{LoggedExerciseID: le.ID, Reps: 5, Weight: 100}
		set.CreatedAt = today.Add(time.Hour + time.Duration(i)*2*time.Minute)
		if err := db.Create(&set).Error; err != nil {
			t.Fatal(err)
		}
	}

	got, err := services.GetWorkoutTimingStats(context.Background(), 30)
	if err != nil {
		t.Fatal(err)
	}
	if got.Days != 30 {
		t.Fatalf("days = %d", got.Days)
	}
	if got.Workout.Count != 1 || got.Workout.AvgSeconds != 240 {
		t.Fatalf("workout = %+v, want 1 x 240s", got.Workout)
	}
	if got.SetInterval.Count != 2 || got.SetInterval.AvgSeconds != 120 {
		t.Fatalf("set interval = %+v", got.SetInterval)
	}
}
