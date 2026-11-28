// Package seed is the e2e suite's baseline data: what the app needs to be usable (the dev login,
// an exercise library with an active program, a food catalogue, a diet plan), nothing a test asserts
// on. Tests create the data they mutate or check.
//
// Reset empties every table and reseeds, so IDs (sequences restart) and dates (relative to Epoch)
// are the same on every run. It is used to build the e2e template database (cmd/e2eseed) and by
// POST /__e2e/reset before each test. It is plain Go with no build tag, but nothing in the server
// imports it outside the e2e build.
package seed

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authm "be-simpletracker/internal/core/auth/models"
	dietm "be-simpletracker/internal/core/diet/models"
	workm "be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/database"
)

// Username and Password of the seeded login (the same as the dev stack's devuser).
const (
	Username = "dev"
	Password = "dev"
)

// Epoch is the suite's default "now": a Thursday at noon, server-local. The e2e fixtures set both
// the server clock and the browser clock to it. Mirrored in e2e/fixtures (E2E_EPOCH).
func Epoch() time.Time { return time.Date(2026, 1, 15, 12, 0, 0, 0, time.Local) }

// EpochDay is local midnight of Epoch, the day boundary the API uses (clock.ZerodTime).
func EpochDay() time.Time {
	e := Epoch()
	return time.Date(e.Year(), e.Month(), e.Day(), 0, 0, 0, 0, time.Local)
}

// Reset truncates every table except schema_migrations (restarting ID sequences) and reseeds,
// in one transaction.
func Reset(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var tables []string
		if err := tx.Raw(`SELECT quote_ident(tablename) FROM pg_tables
			WHERE schemaname = 'public' AND tablename <> 'schema_migrations' ORDER BY tablename`).
			Scan(&tables).Error; err != nil {
			return err
		}
		if len(tables) > 0 {
			if err := tx.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE").Error; err != nil {
				return fmt.Errorf("truncate: %w", err)
			}
		}
		return seed(tx)
	})
}

var (
	hashOnce sync.Once
	hash     string
	hashErr  error
)

// passwordHash is computed once per process at bcrypt.MinCost: Reset runs before every test.
func passwordHash() (string, error) {
	hashOnce.Do(func() {
		b, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.MinCost)
		hash, hashErr = string(b), err
	})
	return hash, hashErr
}

func seed(tx *gorm.DB) error {
	for _, f := range []func(*gorm.DB) error{seedUser, seedGym, seedDiet} {
		if err := f(tx); err != nil {
			return err
		}
	}
	return nil
}

func seedUser(tx *gorm.DB) error {
	h, err := passwordHash()
	if err != nil {
		return err
	}
	birth := 1995
	return tx.Create(&authm.User{
		Username:     Username,
		Password:     h,
		Email:        Username + "@localhost",
		BirthYear:    &birth,
		IsAdmin:      true,
		TokenVersion: authm.InitialTokenVersion,
	}).Error
}

func seedGym(tx *gorm.DB) error {
	owner, err := devUser(tx)
	if err != nil {
		return err
	}
	o := owner
	defs := []struct {
		name string
		lt   workm.ExerciseLoadType
	}{
		{"Barbell Bench Press", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Overhead Press", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Incline Dumbbell Press", workm.ExerciseLoadTypeFreeWeights},
		{"Triceps Pushdown", workm.ExerciseLoadTypeWeightStack},
		{"Barbell Row", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Lat Pulldown", workm.ExerciseLoadTypeWeightStack},
		{"Dumbbell Curl", workm.ExerciseLoadTypeFreeWeights},
		{"Face Pull", workm.ExerciseLoadTypeWeightStack},
		{"Back Squat", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Romanian Deadlift", workm.ExerciseLoadTypePlateLoadedWithBar},
		{"Leg Press", workm.ExerciseLoadTypePlateLoadedWithoutBar},
		{"Leg Curl", workm.ExerciseLoadTypeWeightStack},
	}
	exs := make([]workm.Exercise, len(defs))
	for i, d := range defs {
		exs[i] = workm.Exercise{Creator: database.Creator{CreatedBy: o.UserID}, Name: d.name, LoadType: d.lt, RepRollover: 12}
	}
	if err := tx.Create(&exs).Error; err != nil {
		return err
	}

	// The app expects exactly one active program.
	prog := workm.WorkoutProgram{Owner: o, Name: "Push Pull Legs", IsActive: true}
	if err := tx.Create(&prog).Error; err != nil {
		return err
	}
	plans := []struct {
		name string
		day  time.Weekday
		idx  []int
	}{
		{"Push", time.Monday, []int{0, 1, 2, 3}},
		{"Pull", time.Wednesday, []int{4, 5, 6, 7}},
		{"Legs", time.Thursday, []int{8, 9, 10, 11}}, // the epoch's weekday: "today" has a plan
	}
	for _, p := range plans {
		day := int(p.day)
		wp := workm.WorkoutPlan{Owner: o, Name: p.name, WorkoutProgramID: &prog.ID, DayOfWeek: &day}
		if err := tx.Create(&wp).Error; err != nil {
			return err
		}
		if err := tx.Create(&workm.WorkoutPlanDay{Owner: o, WorkoutPlanID: wp.ID, DayOfWeek: day}).Error; err != nil {
			return err
		}
		for order, ei := range p.idx {
			if err := tx.Create(&workm.WorkoutPlanExercise{Owner: o, WorkoutPlanID: wp.ID, ExerciseID: exs[ei].ID, DisplayOrder: order}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedDiet(tx *gorm.DB) error {
	owner, err := devUser(tx)
	if err != nil {
		return err
	}
	added := database.Creator{CreatedBy: owner.UserID}
	defs := []struct {
		name, serving            string
		amt, cal, p, fib, c, fat float32
	}{
		{"Chicken Breast", "g", 100, 165, 31, 0, 0, 3.6},
		{"White Rice (cooked)", "g", 100, 130, 2.7, 0.4, 28, 0.3},
		{"Egg", "each", 1, 72, 6.3, 0, 0.4, 4.8},
		{"Oats", "g", 40, 150, 5, 4, 27, 3},
		{"Banana", "each", 1, 105, 1.3, 3.1, 27, 0.4},
		{"Greek Yogurt", "g", 170, 100, 17, 0, 6, 0.7},
		{"Broccoli", "g", 100, 34, 2.8, 2.6, 7, 0.4},
		{"Olive Oil", "tbsp", 1, 119, 0, 0, 0, 13.5},
	}
	foods := make([]dietm.Food, len(defs))
	for i, d := range defs {
		foods[i] = dietm.Food{Creator: added, Name: d.name, ServingType: d.serving, ServingAmount: d.amt, Calories: d.cal, Protein: d.p, Fiber: d.fib, Carbs: d.c, Fat: d.fat}
	}
	if err := tx.Create(&foods).Error; err != nil {
		return err
	}
	from := EpochDay().AddDate(0, 0, -60)
	return tx.Create(&dietm.Plan{Owner: owner, Name: "Maintenance", Calories: 2500, Protein: 180, Fiber: 35, Carbs: 280, Fat: 70, EffectiveFrom: &from}).Error
}
