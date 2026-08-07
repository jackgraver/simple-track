package services_test

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/utils"
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestCreateExercise_andGetAllExercises(t *testing.T) {
	testutil.SetupTestDB(t)
	created, err := services.CreateExercise("Bench Press", 12, "squeeze")
	if err != nil {
		t.Fatal(err)
	}
	all, err := services.GetAllExercises(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("got %+v", all)
	}
	excluded, err := services.GetAllExercises([]uint{created.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(excluded) != 0 {
		t.Fatalf("expected empty list, got %+v", excluded)
	}
	if created.LoadType != models.ExerciseLoadTypePlateLoadedWithBar {
		t.Fatalf("expected default load type, got %q", created.LoadType)
	}
}

func TestUpdateExercise_andUpdateExerciseCues(t *testing.T) {
	testutil.SetupTestDB(t)
	created, err := services.CreateExercise("Row", 10, "old")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := services.UpdateExercise(
		created.ID,
		"Barbell Row",
		8,
		"pull elbows back",
		models.ExerciseLoadTypePlateLoadedWithoutBar,
	)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Barbell Row" ||
		updated.RepRollover != 8 ||
		updated.LoadType != models.ExerciseLoadTypePlateLoadedWithoutBar {
		t.Fatalf("got %+v", updated)
	}
	updated, err = services.UpdateExercise(
		created.ID,
		"Barbell Row",
		8,
		"pull elbows back",
		models.ExerciseLoadTypeFreeWeights,
	)
	if err != nil {
		t.Fatal(err)
	}
	if updated.LoadType != models.ExerciseLoadTypeFreeWeights {
		t.Fatalf("expected free weights load type, got %q", updated.LoadType)
	}
	cued, err := services.UpdateExerciseCues(created.ID, "new cue")
	if err != nil {
		t.Fatal(err)
	}
	if cued.Cues != "new cue" {
		t.Fatalf("got %+v", cued)
	}
}

func TestListExercises_paginates(t *testing.T) {
	testutil.SetupTestDB(t)
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		if _, err := services.CreateExercise(name, 10, ""); err != nil {
			t.Fatal(err)
		}
	}
	res, err := services.ListExercises(1, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 3 || len(res.Exercises) != 2 {
		t.Fatalf("got total=%d len=%d", res.Total, len(res.Exercises))
	}
}

func TestLogExercise_UpdateLoggedExercise_DeleteLoggedSet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	ex, err := services.CreateExercise("Curl", 12, "")
	if err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	le := models.LoggedExercise{WorkoutLogID: wl.ID, ExerciseID: ex.ID, Sets: []models.LoggedSet{{Reps: 10, Weight: 20}}}
	if err := services.LogExercise(&le); err != nil {
		t.Fatal(err)
	}
	if len(le.Sets) != 1 || le.Sets[0].ID == 0 {
		t.Fatalf("sets not persisted: %+v", le.Sets)
	}
	setID := le.Sets[0].ID
	le.Sets[0].Reps = 12
	le.Sets[0].Weight = 22.5
	if err := services.UpdateLoggedExercise(le); err != nil {
		t.Fatal(err)
	}
	loaded, err := services.LoadLoggedExercise(le.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Sets[0].Reps != 12 || loaded.Sets[0].Weight != 22.5 {
		t.Fatalf("update failed: %+v", loaded.Sets)
	}
	if err := services.DeleteLoggedSet(setID); err != nil {
		t.Fatal(err)
	}
	_, err = services.LoadLoggedExercise(le.ID)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected logged exercise removed, got %v", err)
	}
}

func TestRemoveLoggedExerciseForDay(t *testing.T) {
	db := testutil.SetupTestDB(t)
	today := utils.ZerodTime(0)
	ex, err := services.CreateExercise("Lat Raise", 15, "")
	if err != nil {
		t.Fatal(err)
	}
	wl := models.WorkoutLog{Date: today}
	if err := db.Create(&wl).Error; err != nil {
		t.Fatal(err)
	}
	le := models.LoggedExercise{WorkoutLogID: wl.ID, ExerciseID: ex.ID}
	if err := services.LogExercise(&le); err != nil {
		t.Fatal(err)
	}
	if err := services.RemoveLoggedExerciseForDay(context.Background(), 0, ex.ID); err != nil {
		t.Fatal(err)
	}
	_, err = services.LoadLoggedExercise(le.ID)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected removal, got %v", err)
	}
}

func TestGetExerciseProgression(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ex, err := services.CreateExercise("OHP", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, day := range []struct {
		offset int
		weight float32
	}{
		{1, 50},
		{0, 52.5},
	} {
		wl := models.WorkoutLog{Date: utils.ZerodTime(day.offset)}
		if err := db.Create(&wl).Error; err != nil {
			t.Fatal(err)
		}
		le := models.LoggedExercise{WorkoutLogID: wl.ID, ExerciseID: ex.ID}
		if err := db.Create(&le).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&models.LoggedSet{LoggedExerciseID: le.ID, Reps: 5, Weight: day.weight}).Error; err != nil {
			t.Fatal(err)
		}
	}
	entries, err := services.GetExerciseProgression(ex.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Weight != 50 || entries[1].Weight != 52.5 {
		t.Fatalf("got %+v", entries)
	}
}
