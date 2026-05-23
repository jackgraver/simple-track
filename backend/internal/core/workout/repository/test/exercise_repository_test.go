package test

import (
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/models"
	"testing"

	"gorm.io/gorm"
)

func TestExerciseExists(t *testing.T) {
	db := setupTestDB(t)
	ex := models.Exercise{Name: "Pull-up"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	if err := workoutrepo.ExerciseExists(ex.ID); err != nil {
		t.Fatal(err)
	}
	if err := workoutrepo.ExerciseExists(9999); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestUpdateExercise_notFound(t *testing.T) {
	setupTestDB(t)
	_, err := workoutrepo.UpdateExercise(9999, "Missing", 10, "")
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestUpdateExerciseCues_notFound(t *testing.T) {
	setupTestDB(t)
	_, err := workoutrepo.UpdateExerciseCues(9999, "cue")
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestFindAllExercises_withExclude(t *testing.T) {
	db := setupTestDB(t)
	for _, name := range []string{"One", "Two"} {
		if err := db.Create(&models.Exercise{Name: name}).Error; err != nil {
			t.Fatal(err)
		}
	}
	var second models.Exercise
	if err := db.Where("name = ?", "Two").First(&second).Error; err != nil {
		t.Fatal(err)
	}
	all, err := workoutrepo.FindAllExercises([]uint{second.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].Name != "One" {
		t.Fatalf("got %+v", all)
	}
}
