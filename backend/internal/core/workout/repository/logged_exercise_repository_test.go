package workoutrepo_test

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/utils"
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestUpdateLoggedExerciseWithSets_addUpdateDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	ex := models.Exercise{Name: "Press"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	le := models.LoggedExercise{
		WorkoutLogID: wl.ID,
		ExerciseID:   ex.ID,
		Sets: []models.LoggedSet{
			{Reps: 8, Weight: 60},
			{Reps: 8, Weight: 60},
		},
	}
	if err := workoutrepo.CreateLoggedExercise(&le); err != nil {
		t.Fatal(err)
	}
	var sets []models.LoggedSet
	if err := db.Where("logged_exercise_id = ?", le.ID).Find(&sets).Error; err != nil {
		t.Fatal(err)
	}
	if len(sets) != 2 {
		t.Fatalf("expected 2 sets, got %d", len(sets))
	}
	sets[0].Reps = 10
	le.Sets = []models.LoggedSet{sets[0], {Reps: 6, Weight: 65}}
	if err := workoutrepo.UpdateLoggedExerciseWithSets(le); err != nil {
		t.Fatal(err)
	}
	var after []models.LoggedSet
	if err := db.Where("logged_exercise_id = ?", le.ID).Order("id ASC").Find(&after).Error; err != nil {
		t.Fatal(err)
	}
	if len(after) != 2 || after[0].Reps != 10 || after[1].Reps != 6 {
		t.Fatalf("got %+v", after)
	}
}

func TestRemoveLoggedExerciseForDay_notFound(t *testing.T) {
	testutil.SetupTestDB(t)
	err := workoutrepo.RemoveLoggedExerciseForDay(context.Background(), utils.ZerodTime(0), 1)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetPreviousExerciseLog_noMatch(t *testing.T) {
	testutil.SetupTestDB(t)
	_, err := workoutrepo.GetPreviousExerciseLog(context.Background(), utils.ZerodTime(0), "Missing", 0)
	if err != nil {
		t.Fatalf("expected nil error with empty result, got %v", err)
	}
}

func TestGetMaxExerciseLog_usesHighestWeightThenReps(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	ex := models.Exercise{Name: "Bench"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	older := models.WorkoutLog{Date: today.AddDate(0, 0, -3)}
	recent := models.WorkoutLog{Date: today.AddDate(0, 0, -1)}
	current := models.WorkoutLog{Date: today}
	if err := db.Create(&older).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&recent).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&current).Error; err != nil {
		t.Fatal(err)
	}
	olderExercise := models.LoggedExercise{WorkoutLogID: older.ID, ExerciseID: ex.ID}
	recentExercise := models.LoggedExercise{WorkoutLogID: recent.ID, ExerciseID: ex.ID}
	currentExercise := models.LoggedExercise{WorkoutLogID: current.ID, ExerciseID: ex.ID}
	if err := db.Create(&olderExercise).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&recentExercise).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&currentExercise).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LoggedSet{LoggedExerciseID: olderExercise.ID, Weight: 100, Reps: 5}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LoggedSet{LoggedExerciseID: recentExercise.ID, Weight: 100, Reps: 8}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.LoggedSet{LoggedExerciseID: currentExercise.ID, Weight: 125, Reps: 1}).Error; err != nil {
		t.Fatal(err)
	}

	got, err := workoutrepo.GetMaxExerciseLog(context.Background(), today, "Bench")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != recentExercise.ID {
		t.Fatalf("got exercise %d want %d", got.ID, recentExercise.ID)
	}
	if len(got.Sets) != 1 || got.Sets[0].Weight != 100 || got.Sets[0].Reps != 8 {
		t.Fatalf("sets %+v", got.Sets)
	}
}
