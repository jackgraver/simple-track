package testutil

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/database"
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Identity is the user module tests act as: the admin, user 1, like the dev login.
var Identity = database.Identity{UserID: 1, IsAdmin: true}

// Ctx is a request-like context for db acting as Identity. SQLite tests have no request
// transaction, so the identity is set by hand.
func Ctx(db *gorm.DB) context.Context {
	return database.WithIdentity(database.WithDB(context.Background(), db), Identity)
}

// SetupTestDB returns an in-memory SQLite database whose statements already carry Identity, so
// rows created straight through it are stamped for user 1 (the Owner and Creator hooks).
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.Exercise{},
		&models.ExerciseSetting{},
		&models.WorkoutProgram{},
		&models.WorkoutPlan{},
		&models.WorkoutPlanDay{},
		&models.WorkoutPlanExercise{},
		&models.WorkoutLog{},
		&models.LoggedExercise{},
		&models.LoggedSet{},
		&models.Cardio{},
	); err != nil {
		t.Fatal(err)
	}
	// The per-user uniques from migration 0017 (the set endpoint's ON CONFLICT needs the first).
	for _, ddl := range []string{
		`CREATE UNIQUE INDEX idx_logged_sets_client_id ON logged_sets (user_id, client_id)`,
		`CREATE UNIQUE INDEX idx_workout_logs_user_date ON workout_logs (user_id, date) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX uq_exercise_settings_user_exercise ON exercise_settings (user_id, exercise_id)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	db = db.WithContext(database.WithIdentity(context.Background(), Identity))
	if err := db.Create(&models.WorkoutProgram{Name: "Test Program", IsActive: true}).Error; err != nil {
		t.Fatal(err)
	}
	return db
}
