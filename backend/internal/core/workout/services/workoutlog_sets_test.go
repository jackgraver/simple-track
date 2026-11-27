package services_test

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"

	"gorm.io/gorm"
)

func setsTestExercise(t *testing.T, db *gorm.DB, name string) models.Exercise {
	t.Helper()
	ex := models.Exercise{Name: name}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	return ex
}

func clientID(n int) string { return fmt.Sprintf("00000000-0000-4000-8000-%012d", n) }

func TestLogSet_createsDayAndParent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	ex := setsTestExercise(t, db, "Squat")

	res, err := services.LogSet(ctx, services.LogSetInput{
		ClientID: clientID(1), Day: clock.ZerodTime(0), ExerciseID: ex.ID, Reps: 5, Weight: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created || !res.FirstSetOfDay {
		t.Fatalf("want created first set, got %+v", res)
	}
	if len(res.Exercise.Sets) != 1 || res.Exercise.ExerciseID != ex.ID || res.Exercise.Exercise == nil {
		t.Fatalf("unexpected exercise in response: %+v", res.Exercise)
	}
	if res.Set.ClientID == nil || *res.Set.ClientID != clientID(1) {
		t.Fatalf("client_id not stored: %+v", res.Set)
	}

	res2, err := services.LogSet(ctx, services.LogSetInput{
		ClientID: clientID(2), Day: clock.ZerodTime(0), ExerciseID: ex.ID, Reps: 5, Weight: 105,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res2.Created || res2.FirstSetOfDay {
		t.Fatalf("second set: want created, not first of day, got %+v", res2)
	}
	if res2.Exercise.ID != res.Exercise.ID || len(res2.Exercise.Sets) != 2 {
		t.Fatalf("second set should join the same parent: %+v", res2.Exercise)
	}
	if res2.Exercise.Sets[0].ID >= res2.Exercise.Sets[1].ID {
		t.Fatalf("sets should be ordered by id")
	}
}

func TestLogSet_offTodayUsesThatDay(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	ex := setsTestExercise(t, db, "Row")
	yesterday := clock.ZerodTime(1)
	res, err := services.LogSet(ctx, services.LogSetInput{
		ClientID: clientID(1), Day: yesterday, ExerciseID: ex.ID, Reps: 8, Weight: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Exercise.LogDate.Equal(yesterday) {
		t.Fatalf("log date = %v, want %v", res.Exercise.LogDate, yesterday)
	}
}

func TestLogSet_replayIsIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	ex := setsTestExercise(t, db, "Bench")
	in := services.LogSetInput{ClientID: clientID(7), Day: clock.ZerodTime(0), ExerciseID: ex.ID, Reps: 5, Weight: 80}

	first, err := services.LogSet(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	// A retry may carry different values (e.g. the user edited while offline); first write wins.
	in.Weight = 90
	again, err := services.LogSet(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if again.Created || again.FirstSetOfDay {
		t.Fatalf("replay should report created=false, got %+v", again)
	}
	if again.Set.ID != first.Set.ID || again.Set.Weight != 80 {
		t.Fatalf("replay returned a different set: %+v vs %+v", again.Set, first.Set)
	}
	var sets, parents int64
	db.Model(&models.LoggedSet{}).Count(&sets)
	db.Model(&models.LoggedExercise{}).Count(&parents)
	if sets != 1 || parents != 1 {
		t.Fatalf("want 1 set and 1 parent, got %d sets, %d parents", sets, parents)
	}
}

func TestLogSet_clientIDReusedForOtherExerciseConflicts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	a := setsTestExercise(t, db, "Curl")
	b := setsTestExercise(t, db, "Press")
	if _, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(3), Day: clock.ZerodTime(0), ExerciseID: a.ID, Reps: 10}); err != nil {
		t.Fatal(err)
	}
	_, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(3), Day: clock.ZerodTime(0), ExerciseID: b.ID, Reps: 10})
	if !errors.Is(err, workoutrepo.ErrClientIDConflict) {
		t.Fatalf("want ErrClientIDConflict, got %v", err)
	}
}

func TestLogSet_unknownExercise(t *testing.T) {
	ctx := database.WithDB(context.Background(), testutil.SetupTestDB(t))
	_, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(1), Day: clock.ZerodTime(0), ExerciseID: 999})
	if !errors.Is(err, services.ErrExerciseNotFound) {
		t.Fatalf("want ErrExerciseNotFound, got %v", err)
	}
}

// SQLite serializes writers, so this checks the logic (no duplicate parent, replays collapse)
// rather than real parallelism; pg_sets_test.go runs the same scenario on Postgres.
func TestLogSet_concurrentFirstSets(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1) // :memory: is per connection
	ex := setsTestExercise(t, db, "Deadlift")
	if _, err := services.GetOrCreateToday(ctx, 0); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 6 distinct sets, each sent twice (a double tap / retry of the same create).
			_, err := services.LogSet(ctx, services.LogSetInput{
				ClientID: clientID(100 + i%6), Day: clock.ZerodTime(0), ExerciseID: ex.ID, Reps: 5, Weight: 140,
			})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var sets, parents int64
	db.Model(&models.LoggedSet{}).Count(&sets)
	db.Model(&models.LoggedExercise{}).Count(&parents)
	if sets != 6 || parents != 1 {
		t.Fatalf("want 6 sets under 1 parent, got %d sets, %d parents", sets, parents)
	}
}

func TestUpdateSet_patchByIDAndClientID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	ex := setsTestExercise(t, db, "Dip")
	res, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(5), Day: clock.ZerodTime(0), ExerciseID: ex.ID, Reps: 8, Weight: 20, WeightSetup: "belt"})
	if err != nil {
		t.Fatal(err)
	}
	reps := uint(10)
	set, exercise, err := services.UpdateSet(ctx, workoutrepo.SetRef{ClientID: clientID(5)}, workoutrepo.SetPatch{Reps: &reps})
	if err != nil {
		t.Fatal(err)
	}
	if set.Reps != 10 || set.Weight != 20 || set.WeightSetup != "belt" {
		t.Fatalf("only reps should change: %+v", set)
	}
	if len(exercise.Sets) != 1 || exercise.Sets[0].Reps != 10 {
		t.Fatalf("response exercise not refreshed: %+v", exercise.Sets)
	}
	w := float32(25)
	set, _, err = services.UpdateSet(ctx, workoutrepo.SetRef{ID: res.Set.ID}, workoutrepo.SetPatch{Weight: &w})
	if err != nil || set.Weight != 25 || set.Reps != 10 {
		t.Fatalf("patch by id: %+v, %v", set, err)
	}
	if _, _, err := services.UpdateSet(ctx, workoutrepo.SetRef{ID: 9999}, workoutrepo.SetPatch{Weight: &w}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing set: want ErrRecordNotFound, got %v", err)
	}
}

func TestDeleteSet_repeatIsSuccess(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	ex := setsTestExercise(t, db, "Lunge")
	a, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(1), Day: clock.ZerodTime(0), ExerciseID: ex.ID, Reps: 8})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(2), Day: clock.ZerodTime(0), ExerciseID: ex.ID, Reps: 8}); err != nil {
		t.Fatal(err)
	}

	res, err := services.DeleteSet(ctx, workoutrepo.SetRef{ID: a.Set.ID})
	if err != nil || !res.Deleted || res.ExerciseRemoved {
		t.Fatalf("first delete: %+v, %v", res, err)
	}
	res, err = services.DeleteSet(ctx, workoutrepo.SetRef{ID: a.Set.ID})
	if err != nil || res.Deleted {
		t.Fatalf("repeat delete should succeed with deleted=false: %+v, %v", res, err)
	}
	// Last set, by client_id: the parent goes too (legacy behaviour).
	res, err = services.DeleteSet(ctx, workoutrepo.SetRef{ClientID: clientID(2)})
	if err != nil || !res.Deleted || !res.ExerciseRemoved || res.LoggedExerciseID != a.Exercise.ID {
		t.Fatalf("last delete: %+v, %v", res, err)
	}
	var parents int64
	db.Model(&models.LoggedExercise{}).Count(&parents)
	if parents != 0 {
		t.Fatalf("parent should be removed with its last set, got %d", parents)
	}
}

func TestLogExercise_secondCreateReportsExisting(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := database.WithDB(context.Background(), db)
	ex := setsTestExercise(t, db, "Shrug")
	wl, err := services.GetOrCreateToday(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	first := models.LoggedExercise{WorkoutLogID: wl.ID, ExerciseID: ex.ID, Sets: []models.LoggedSet{{Reps: 5}}}
	if err := services.LogExercise(ctx, &first); err != nil {
		t.Fatal(err)
	}
	second := models.LoggedExercise{WorkoutLogID: wl.ID, ExerciseID: ex.ID, Sets: []models.LoggedSet{{Reps: 5}}}
	err = services.LogExercise(ctx, &second)
	if !errors.Is(err, workoutrepo.ErrLoggedExerciseExists) || second.ID != first.ID {
		t.Fatalf("want ErrLoggedExerciseExists pointing at %d, got %v (id %d)", first.ID, err, second.ID)
	}
	var sets int64
	db.Model(&models.LoggedSet{}).Count(&sets)
	if sets != 1 {
		t.Fatalf("conflicting create must not add sets, got %d", sets)
	}
}
