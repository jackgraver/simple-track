package models

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type WorkoutModel struct {
	db *gorm.DB
}

func NewWorkoutModel(db *gorm.DB) *WorkoutModel {
	return &WorkoutModel{db: db}
}

func (m *WorkoutModel) MigrateDatabase() {
	fmt.Println("Migrating workout database")
    if err := m.db.Migrator().DropTable(
        &WorkoutPlan{},
        &PlannedExercise{},
        &PlannedSet{},
        &WorkoutLog{},
        &LoggedExercise{},
        &LoggedSet{},
        &Cardio{},
    ); err != nil {
		fmt.Printf("Failed to drop workout database: %v\n", err)
	}

    if err := m.db.AutoMigrate(
        &WorkoutPlan{},
        &PlannedExercise{},
        &PlannedSet{},
        &WorkoutLog{},
        &LoggedExercise{},
        &LoggedSet{},
        &Cardio{},
    ); err != nil {
		fmt.Printf("Failed to migrate workout database: %v\n", err)
	}

	if err := m.seedDatabase(); err != nil {
		fmt.Printf("Failed to seed workout database: %v\n", err)
	}
}

func (m *WorkoutModel) seedDatabase() error {
	fmt.Println("Seeding workout database")

	push := WorkoutPlan{
		Name: "Push",
		Exercises: []PlannedExercise{
			{
				Name: "Incline Press",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Chest Fly",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Dips",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Lat Raise",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Shoulder Press",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "JM Press",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Extensions",
				Sets: []PlannedSet{
				},
			},
		},
	}
	pull := WorkoutPlan{
		Name: "Pull",
		Exercises: []PlannedExercise{
			{
				Name: "Barbell Rows",
				Sets: []PlannedSet{
				},
			},	
			{
				Name: "Face Pulls",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Pulldowns",
				Sets: []PlannedSet{
				},
			},		
			{
				Name: "Incline Curls",
				Sets: []PlannedSet{
				},
			},			
			{
				Name: "Hammer Curls",
				Sets: []PlannedSet{
				},
			},
		},
	}
	legs := WorkoutPlan{
		Name: "Legs",
		Exercises: []PlannedExercise{
			{
				Name: "Outer Thigh",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Inner Thigh",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Leg Extensions",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Hamstring Cruls",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Squat",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Deadlift",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Calf Raises",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Ab Crunches",
				Sets: []PlannedSet{
				},
			},
		},
	}

	upper := WorkoutPlan{
		Name: "Upper",
		Exercises: []PlannedExercise{
			{
				Name: "Incline Press",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Chest Fly",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Dips",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Lat Raise",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Shoulder Press",
				Sets: []PlannedSet{
				},
			},
						{
				Name: "Barbell Rows",
				Sets: []PlannedSet{
				},
			},	
			{
				Name: "Face Pulls",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Pulldowns",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "JM Press",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Extensions",
				Sets: []PlannedSet{
				},
			},		
			{
				Name: "Incline Curls",
				Sets: []PlannedSet{
				},
			},			
			{
				Name: "Hammer Curls",
				Sets: []PlannedSet{
				},
			},
		},
	}

	lower := WorkoutPlan{
		Name: "Lower",
		Exercises: []PlannedExercise{
			{
				Name: "Outer Thigh",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Inner Thigh",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Leg Extensions",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Hamstring Cruls",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Squat",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Deadlift",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Hip Hinge",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Leg Press",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Calf Raises",
				Sets: []PlannedSet{
				},
			},
			{
				Name: "Ab Crunches",
				Sets: []PlannedSet{
				},
			},
		},
	}

	active_rest := WorkoutPlan{
		Name: "Active Rest",
		Exercises: []PlannedExercise{
			{
				Name: "Abs",
				Sets: []PlannedSet{
				},
			},
		},
	}

	rest := WorkoutPlan{
		Name: "Rest",
		Exercises: []PlannedExercise{},
	}

	m.db.Create(&push)
	m.db.Create(&pull)
	m.db.Create(&legs)
	m.db.Create(&upper)
	m.db.Create(&lower)
	m.db.Create(&active_rest)
	m.db.Create(&rest)

	year := 2025
	now := time.Now()
	start := time.Date(year, time.September, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, now.Location())

	// Map weekday → workout plan
	weekdayPlans := map[time.Weekday]WorkoutPlan{
		time.Sunday:    lower,
		time.Monday:    rest,
		time.Tuesday:   push,
		time.Wednesday: pull,
		time.Thursday:  legs,
		time.Friday:    active_rest,
		time.Saturday:  upper,	
	}

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		plan := weekdayPlans[date.Weekday()]

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

		// Add treadmill on leg days
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

		m.db.Create(&wl)
	}
	return nil;
}

func (m *WorkoutModel) Preloads() []string {
	return []string{"WorkoutPlan.Exercises.Sets", "Cardio", "Exercises.Sets"}
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
    WorkoutPlan   *WorkoutPlan `json:"workout_plan"` //TODO deprecate plan other than unique name? no exercise tracking because we can just look at previous weeks
    Exercises []LoggedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
    Cardio    *Cardio          `json:"cardio" gorm:"constraint:OnDelete:CASCADE;"`
}

func (w *WorkoutLog) Preloads() []string {
    return []string{"WorkoutPlan.Exercises.Sets", "Cardio", "Exercises.Sets"}
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

