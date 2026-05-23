package test

import (
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/utils"
	"testing"

	"gorm.io/gorm"
)

func TestDeleteLoggedSet_removesOrphanExercise(t *testing.T) {
	db := setupTestDB(t)
	today := utils.ZerodTime(0)
	ex := models.Exercise{Name: "Fly"}
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
	set := models.LoggedSet{LoggedExerciseID: le.ID, Reps: 12, Weight: 15}
	if err := db.Create(&set).Error; err != nil {
		t.Fatal(err)
	}
	if err := workoutrepo.DeleteLoggedSet(set.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&models.LoggedSet{}, set.ID).Error; err != gorm.ErrRecordNotFound {
		t.Fatalf("set should be deleted, got %v", err)
	}
	if err := db.First(&models.LoggedExercise{}, le.ID).Error; err != gorm.ErrRecordNotFound {
		t.Fatalf("logged exercise should be deleted, got %v", err)
	}
}

func TestDeleteLoggedSet_notFound(t *testing.T) {
	setupTestDB(t)
	err := workoutrepo.DeleteLoggedSet(9999)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestDeleteLoggedSet_keepsExerciseWhenSetsRemain(t *testing.T) {
	db := setupTestDB(t)
	le := models.LoggedExercise{WorkoutLogID: 1, ExerciseID: 1}
	if err := db.Create(&le).Error; err != nil {
		t.Fatal(err)
	}
	set1 := models.LoggedSet{LoggedExerciseID: le.ID, Reps: 10, Weight: 20}
	set2 := models.LoggedSet{LoggedExerciseID: le.ID, Reps: 10, Weight: 20}
	if err := db.Create(&set1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&set2).Error; err != nil {
		t.Fatal(err)
	}
	if err := workoutrepo.DeleteLoggedSet(set1.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&models.LoggedExercise{}, le.ID).Error; err != nil {
		t.Fatalf("exercise should remain, got %v", err)
	}
}
