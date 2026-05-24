package workoutrepo_test

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/testutil"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestReorderPlanExercises_rejectsPartialList(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := models.WorkoutPlan{Name: "Split"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	ex1 := models.Exercise{Name: "A"}
	ex2 := models.Exercise{Name: "B"}
	if err := db.Create(&ex1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&ex2).Error; err != nil {
		t.Fatal(err)
	}
	for i, ex := range []models.Exercise{ex1, ex2} {
		if err := db.Create(&models.WorkoutPlanExercise{
			WorkoutPlanID: plan.ID,
			ExerciseID:    ex.ID,
			DisplayOrder:  i,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	err := workoutrepo.ReorderPlanExercises(plan.ID, []uint{ex1.ID})
	if err == nil || !strings.Contains(err.Error(), "must include all plan exercises") {
		t.Fatalf("expected validation error, got %v", err)
	}
	err = workoutrepo.ReorderPlanExercises(plan.ID, []uint{ex1.ID, 9999})
	if err == nil || !strings.Contains(err.Error(), "invalid exercise id") {
		t.Fatalf("expected invalid id error, got %v", err)
	}
}

func TestRemoveExerciseFromPlan_notFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := models.WorkoutPlan{Name: "Empty"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	err := workoutrepo.RemoveExerciseFromPlan(plan.ID, 1)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestAddExerciseToPlan_rejectsMissingPlan(t *testing.T) {
	testutil.SetupTestDB(t)
	err := workoutrepo.AddExerciseToPlan(9999, 1)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestLoadPlanWithOrderedExercises_notFound(t *testing.T) {
	testutil.SetupTestDB(t)
	_, err := workoutrepo.LoadPlanWithOrderedExercises(9999)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestRenumberPlanExerciseDisplayOrder_onRemove(t *testing.T) {
	db := testutil.SetupTestDB(t)
	plan := models.WorkoutPlan{Name: "Three"}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	var exercises []models.Exercise
	for _, name := range []string{"A", "B", "C"} {
		ex := models.Exercise{Name: name}
		if err := db.Create(&ex).Error; err != nil {
			t.Fatal(err)
		}
		exercises = append(exercises, ex)
	}
	for i, ex := range exercises {
		if err := db.Create(&models.WorkoutPlanExercise{
			WorkoutPlanID: plan.ID,
			ExerciseID:    ex.ID,
			DisplayOrder:  i,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := workoutrepo.RemoveExerciseFromPlan(plan.ID, exercises[0].ID); err != nil {
		t.Fatal(err)
	}
	ordered, err := workoutrepo.LoadExercisesOrderedForPlan(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(ordered) != 2 || ordered[0].Name != "B" || ordered[1].Name != "C" {
		t.Fatalf("got %+v", ordered)
	}
}
