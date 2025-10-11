package workout

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
)
func MigrateWorkoutDatabase(db *gorm.DB) {
    db.Migrator().DropTable(
        &WorkoutPlan{},
        &PlannedExercise{},
        &PlannedSet{},
        &WorkoutLog{},
        &LoggedExercise{},
        &LoggedSet{},
        &Cardio{},
    )

    db.AutoMigrate(
        &WorkoutPlan{},
        &PlannedExercise{},
        &PlannedSet{},
        &WorkoutLog{},
        &LoggedExercise{},
        &LoggedSet{},
        &Cardio{},
    )
	seed(db)
}

func seed(db *gorm.DB) {
	push := WorkoutPlan{
		Name: "Push Day",
		Exercises: []PlannedExercise{
			{
				Name: "Bench Press",
				Sets: []PlannedSet{
					{Reps: 10, Weight: 100},
					{Reps: 10, Weight: 100},
					{Reps: 8, Weight: 105},
					{Reps: 6, Weight: 110},
				},
			},
			{
				Name: "Overhead Press",
				Sets: []PlannedSet{
					{Reps: 8, Weight: 60},
					{Reps: 8, Weight: 60},
					{Reps: 6, Weight: 65},
				},
			},
			{
				Name: "Tricep Pushdown",
				Sets: []PlannedSet{
					{Reps: 12, Weight: 45},
					{Reps: 12, Weight: 45},
					{Reps: 12, Weight: 45},
				},
			},
		},
	}
	pull := WorkoutPlan{
		Name: "Pull Day",
		Exercises: []PlannedExercise{
			{
				Name: "Pull-ups",
				Sets: []PlannedSet{
					{Reps: 10, Weight: 0},
					{Reps: 8, Weight: 0},
					{Reps: 6, Weight: 0},
				},
			},
			{
				Name: "Barbell Row",
				Sets: []PlannedSet{
					{Reps: 10, Weight: 95},
					{Reps: 10, Weight: 100},
					{Reps: 8, Weight: 105},
				},
			},
			{
				Name: "Bicep Curl",
				Sets: []PlannedSet{
					{Reps: 12, Weight: 25},
					{Reps: 12, Weight: 25},
					{Reps: 10, Weight: 30},
				},
			},
		},
	}
	legs := WorkoutPlan{
		Name: "Leg Day",
		Exercises: []PlannedExercise{
			{
				Name: "Squat",
				Sets: []PlannedSet{
					{Reps: 8, Weight: 135},
					{Reps: 8, Weight: 155},
					{Reps: 6, Weight: 175},
				},
			},
			{
				Name: "Deadlift",
				Sets: []PlannedSet{
					{Reps: 5, Weight: 185},
					{Reps: 5, Weight: 205},
					{Reps: 3, Weight: 225},
				},
			},
			{
				Name: "Lunges",
				Sets: []PlannedSet{
					{Reps: 12, Weight: 40},
					{Reps: 12, Weight: 40},
				},
			},
		},
	}

	db.Create(&push)
	db.Create(&pull)
	db.Create(&legs)

	year := 2025
	start := time.Date(year, time.September, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)

	plans := []WorkoutPlan{push, pull, legs}
	planIndex := 0

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
        if date.Weekday() == time.Sunday {
            continue
        }

		plan := plans[planIndex%len(plans)]

		var loggedExercises []LoggedExercise
		for _, pe := range plan.Exercises {
			var loggedSets []LoggedSet
			for _, ps := range pe.Sets {
				loggedSets = append(loggedSets, LoggedSet{
					Reps:   ps.Reps + (rand.Intn(3) - 1),
					Weight: ps.Weight + float32((rand.Intn(3)-1)*5),
				})
			}
			loggedExercises = append(loggedExercises, LoggedExercise{
				Name: pe.Name,
				Sets: loggedSets,
			})
		}

		if plan.Name == "Leg Day" {
			loggedExercises = append(loggedExercises, LoggedExercise{
				Name: "Treadmill",
			})
		}

        cardio := Cardio{
            Minutes: 20 + rand.Intn(15),
            Type:    "Running",
        }

		wl := WorkoutLog{
			WorkoutPlanID: &plan.ID,
			Date:          date,
            Cardio:        &cardio,
			Exercises:     loggedExercises,
		}
		db.Create(&wl)

		planIndex++
	}
}

type WorkoutPlan struct {
    gorm.Model
    Name      string            `json:"name"`
    Exercises []PlannedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
}

type PlannedExercise struct {
    gorm.Model
    WorkoutPlanID uint           `json:"workout_plan_id"`
    Name          string         `json:"name"`
    Sets          []PlannedSet   `json:"sets" gorm:"constraint:OnDelete:CASCADE;"`
}

type PlannedSet struct {
    gorm.Model
    PlannedExerciseID uint   `json:"planned_exercise_id"`
    Reps              int    `json:"reps"`
    Weight            float32 `json:"weight"`
}

type WorkoutLog struct {
    gorm.Model
    Date          time.Time    `json:"date"`
    WorkoutPlanID *uint        `json:"workout_plan_id"`
    WorkoutPlan   *WorkoutPlan `json:"workout_plan"`
    Exercises []LoggedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
    Cardio    *Cardio          `json:"cardio" gorm:"constraint:OnDelete:CASCADE;"`
}

type LoggedExercise struct {
    gorm.Model
    WorkoutLogID uint         `json:"workout_log_id"`
    Name         string       `json:"name"`
    Sets         []LoggedSet  `json:"sets" gorm:"constraint:OnDelete:CASCADE;"`
}

type LoggedSet struct {
    gorm.Model
    LoggedExerciseID uint   `json:"logged_exercise_id"`
    Reps             int    `json:"reps"`
    Weight           float32 `json:"weight"`
}

type Cardio struct {
    gorm.Model
    WorkoutLogID uint   `json:"workout_log_id" gorm:"uniqueIndex;not null"`
    Minutes      int    `json:"minutes"`
    Type         string `json:"type"`
}

