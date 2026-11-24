package services_test

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/core/workout/testutil"
	"be-simpletracker/internal/utils"
	"context"
	"errors"
	"testing"
)

func TestExerciseWeights_createUpdateAndClear(t *testing.T) {
	testutil.SetupTestDB(t)
	created, err := services.CreateExerciseWith(services.ExerciseFields{
		Name:        "EZ Curl",
		RepRollover: 12,
		LoadType:    models.ExerciseLoadTypePlateLoadedPerSide,
		BaseWeight:  models.Some(float32(25)),
		WeightStep:  models.Some(float32(5)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.BaseWeight == nil || *created.BaseWeight != 25 || created.WeightStep == nil || *created.WeightStep != 5 {
		t.Fatalf("got %+v", created)
	}

	// Absent weights are left alone.
	updated, err := services.UpdateExerciseWith(created.ID, services.ExerciseFields{Name: "EZ Bar Curl", RepRollover: 12})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "EZ Bar Curl" || updated.LoadType != models.ExerciseLoadTypePlateLoadedPerSide ||
		updated.BaseWeight == nil || *updated.BaseWeight != 25 || updated.WeightStep == nil || *updated.WeightStep != 5 {
		t.Fatalf("absent fields should keep their values: %+v", updated)
	}

	// Null clears back to the load type's default.
	updated, err = services.UpdateExerciseWith(created.ID, services.ExerciseFields{
		Name:        "EZ Bar Curl",
		RepRollover: 12,
		WeightStep:  models.Optional[float32]{Set: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.WeightStep != nil || updated.BaseWeight == nil {
		t.Fatalf("null step should clear only the step: %+v", updated)
	}
}

func TestExerciseWeights_legacyLoadTypes(t *testing.T) {
	testutil.SetupTestDB(t)
	withBar, err := services.CreateExercise("Squat", 10, "", models.ExerciseLoadTypePlateLoadedWithBar)
	if err != nil {
		t.Fatal(err)
	}
	if withBar.LoadType != models.ExerciseLoadTypePlateLoadedPerSide || withBar.BaseWeight == nil || *withBar.BaseWeight != 45 {
		t.Fatalf("with_bar: %+v", withBar)
	}
	// An explicit base weight wins over the legacy type's.
	sled, err := services.CreateExerciseWith(services.ExerciseFields{
		Name:       "Leg Press",
		LoadType:   models.ExerciseLoadTypePlateLoadedWithoutBar,
		BaseWeight: models.Some(float32(90)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if sled.BaseWeight == nil || *sled.BaseWeight != 90 {
		t.Fatalf("explicit base: %+v", sled)
	}
}

func TestExerciseWeights_rejectsBadValues(t *testing.T) {
	testutil.SetupTestDB(t)
	for _, f := range []services.ExerciseFields{
		{Name: "A", BaseWeight: models.Some(float32(-5))},
		{Name: "B", WeightStep: models.Some(float32(0))},
		{Name: "C", WeightStep: models.Some(float32(-2.5))},
	} {
		if _, err := services.CreateExerciseWith(f); !errors.Is(err, services.ErrInvalidExercise) {
			t.Fatalf("%s: want ErrInvalidExercise, got %v", f.Name, err)
		}
	}
}

func TestPlateSetup_logPatchAndClear(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ex := setsTestExercise(t, db, "Bench")
	ctx := context.Background()
	setup := &models.PlateSetup{Base: 45, Mode: models.PlateModePerSide, Plates: map[string]int{"45": 1, "2.5": 1}}
	res, err := services.LogSet(ctx, services.LogSetInput{
		ClientID: clientID(40), Day: utils.ZerodTime(0), ExerciseID: ex.ID, Reps: 5, Weight: 140, PlateSetup: setup,
	})
	if err != nil {
		t.Fatal(err)
	}
	var stored models.LoggedSet
	if err := db.First(&stored, res.Set.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.PlateSetup == nil || stored.PlateSetup.Base != 45 || stored.PlateSetup.Plates["2.5"] != 1 {
		t.Fatalf("stored plate setup: %+v", stored.PlateSetup)
	}

	// A patch without plate_setup leaves it alone.
	reps := uint(6)
	set, _, err := services.UpdateSet(ctx, workoutrepo.SetRef{ClientID: clientID(40)}, workoutrepo.SetPatch{Reps: &reps})
	if err != nil || set.PlateSetup == nil || set.PlateSetup.Plates["45"] != 1 {
		t.Fatalf("reps-only patch: %+v, %v", set.PlateSetup, err)
	}

	// A new setup replaces it.
	next := models.PlateSetup{Base: 45, Mode: models.PlateModePerSide, Plates: map[string]int{"45": 1, "5": 1}}
	set, _, err = services.UpdateSet(ctx, workoutrepo.SetRef{ClientID: clientID(40)}, workoutrepo.SetPatch{PlateSetup: models.Some(next)})
	if err != nil || set.PlateSetup == nil || set.PlateSetup.Plates["5"] != 1 || set.PlateSetup.Plates["2.5"] != 0 || set.Reps != 6 {
		t.Fatalf("replace: %+v, %v", set, err)
	}

	// Null clears it.
	set, _, err = services.UpdateSet(ctx, workoutrepo.SetRef{ClientID: clientID(40)}, workoutrepo.SetPatch{PlateSetup: models.Optional[models.PlateSetup]{Set: true}})
	if err != nil || set.PlateSetup != nil {
		t.Fatalf("clear: %+v, %v", set.PlateSetup, err)
	}
}

func TestPlateSetup_validate(t *testing.T) {
	ok := models.PlateSetup{Base: 45, Mode: models.PlateModePerSide, Plates: map[string]int{"45": 2, "2.5": 1}}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string]models.PlateSetup{
		"mode":      {Mode: "both"},
		"base":      {Mode: models.PlateModeTotal, Base: -1},
		"key":       {Mode: models.PlateModeTotal, Plates: map[string]int{"heavy": 1}},
		"zero key":  {Mode: models.PlateModeTotal, Plates: map[string]int{"0": 1}},
		"neg count": {Mode: models.PlateModeTotal, Plates: map[string]int{"45": -1}},
	} {
		if bad.Validate() == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
