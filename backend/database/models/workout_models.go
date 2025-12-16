package models

import (
	"fmt"
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
        &WorkoutLog{},
        &LoggedExercise{},
		&Exercise{},
        &LoggedSet{},
        &Cardio{},
    ); err != nil {
		fmt.Printf("Failed to drop workout database: %v\n", err)
	}

    if err := m.db.AutoMigrate(
        &WorkoutPlan{},
        &WorkoutLog{},
        &LoggedExercise{},
		&Exercise{},
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

	inclinePress := Exercise{Name: "Incline Press", RepRollover: 10}
	chestFly := Exercise{Name: "Chest Fly", RepRollover: 10}
	dips := Exercise{Name: "Dips", RepRollover: 10}
	latRaise := Exercise{Name: "Lat Raise", RepRollover: 10}
	shoulderPress := Exercise{Name: "Shoulder Press", RepRollover: 10}
	squat := Exercise{Name: "Squat", RepRollover: 10}
	deadlift := Exercise{Name: "Deadlift", RepRollover: 10}
	benchPress := Exercise{Name: "Bench Press", RepRollover: 10}
	cableRows := Exercise{Name: "Cable Rows", RepRollover: 10}
	barbellRows := Exercise{Name: "Barbell Rows", RepRollover: 10}
	facePulls := Exercise{Name: "Face Pulls", RepRollover: 10}
	pulldowns := Exercise{Name: "Pulldowns", RepRollover: 10}
	JMPress := Exercise{Name: "JM Press", RepRollover: 10}
	extensions := Exercise{Name: "Extensions", RepRollover: 10}
	inclineCurls := Exercise{Name: "Incline Curls", RepRollover: 10}
	hammerCurls := Exercise{Name: "Hammer Curls", RepRollover: 10}
	calfRaise := Exercise{Name: "Calf Raises", RepRollover: 10}
	abCrunches := Exercise{Name: "Ab Crunches", RepRollover: 10}
	legPress := Exercise{Name: "Leg Press", RepRollover: 10}
	legExtensions := Exercise{Name: "Leg Extensions", RepRollover: 10}
	hamstringCurls := Exercise{Name: "Hamstring Curls", RepRollover: 10}
	hipPress := Exercise{Name: "Hip Press", RepRollover: 10}
	hipExtensions := Exercise{Name: "Hip Extensions", RepRollover: 10}
	outerThigh := Exercise{Name: "Outer Thigh", RepRollover: 10}
	innerThigh := Exercise{Name: "Inner Thigh", RepRollover: 10}
	hackSquat := Exercise{Name: "Hack Squat", RepRollover: 10}

	exercises := []*Exercise{
		&inclinePress,
		&chestFly,
		&dips,
		&latRaise,
		&shoulderPress,
		&squat,
		&deadlift,
		&benchPress,
		&cableRows,
		&barbellRows,
		&facePulls,
		&pulldowns,
		&JMPress,
		&extensions,
		&inclineCurls,
		&hammerCurls,
		&calfRaise,
		&abCrunches,
		&legPress,
		&legExtensions,
		&hamstringCurls,
		&hipPress,
		&hipExtensions,
		&outerThigh,
		&innerThigh,
	}

	for _, exercise := range exercises {
		m.db.Create(exercise)
	}

	push_plan := WorkoutPlan{
		Name: "Push",
		Exercises: []Exercise{
			inclinePress,
			chestFly,
			dips,
			latRaise,
			shoulderPress,
			JMPress,
			extensions,
		},
	}
	pull_plan := WorkoutPlan{
		Name: "Pull",
		Exercises: []Exercise{
			barbellRows,
			facePulls,
			pulldowns,
			cableRows,
			inclineCurls,
			hammerCurls,
		},
	}
	legs_plan := WorkoutPlan{
		Name: "Legs",
		Exercises: []Exercise{
			outerThigh,
			innerThigh,
			legExtensions,
			hamstringCurls,
			squat,
			deadlift,
			calfRaise,
		},
	}
	upper_plan := WorkoutPlan{
		Name: "Upper",
		Exercises: []Exercise{
			inclinePress,
			chestFly,
			dips,
			latRaise,
			shoulderPress,
			barbellRows,
			pulldowns,
			JMPress,
			extensions,
			inclineCurls,	
			hammerCurls,
		},
	}
	lower_plan := WorkoutPlan{
		Name: "Lower",
		Exercises: []Exercise{
			outerThigh,
			innerThigh,
			legExtensions,
			hamstringCurls,
			squat,
			deadlift,
			hackSquat,
			calfRaise,
		},
	}
	rest_plan := WorkoutPlan{
		Name: "Rest",
	}
	
	m.db.Create(&push_plan)
	m.db.Create(&pull_plan)
	m.db.Create(&legs_plan)
	m.db.Create(&upper_plan)
	m.db.Create(&lower_plan)
	m.db.Create(&rest_plan)

	now := time.Now()
	year := 2025
	start := time.Date(year, time.September, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, now.Location())

	weekdayPlans := map[time.Weekday]WorkoutPlan{
		time.Sunday:    lower_plan,
		time.Monday:    rest_plan,
		time.Tuesday:   push_plan,
		time.Wednesday: pull_plan,
		time.Thursday:  legs_plan,
		time.Friday:    rest_plan,
		time.Saturday:  upper_plan,	
	}

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		plan := weekdayPlans[date.Weekday()]
		wl := WorkoutLog{
			Date: date,
			WorkoutPlan: &plan,
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
    Exercises []Exercise  `gorm:"many2many:workout_plan_exercises;" json:"exercises"`
}

type WorkoutLog struct {
    gorm.Model
    Date          time.Time    `json:"date"`
    WorkoutPlanID *uint        `json:"workout_plan_id"`
    WorkoutPlan   *WorkoutPlan `json:"workout_plan"`
    Exercises []LoggedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
    Cardio    *Cardio          `json:"cardio" gorm:"constraint:OnDelete:CASCADE;"`
}

func (w *WorkoutLog) Preloads() []string {
    return []string{"Cardio", "Exercises.Sets"}
}

type LoggedExercise struct {
    gorm.Model
    WorkoutLogID uint         `json:"workout_log_id"`
    ExerciseID   uint         `json:"exercise_id"`
	Exercise     *Exercise    `json:"exercise"`
    Sets         []LoggedSet  `json:"sets" gorm:"constraint:OnDelete:CASCADE;"`
	Notes        string  	  `json:"notes"`
	PercentChange float32     `json:"percent_change" gorm:"-"`
}

type LoggedSet struct {
    gorm.Model
    LoggedExerciseID uint    `json:"logged_exercise_id"`
    Reps             uint     `json:"reps"`
    Weight           float32 `json:"weight"`
    WeightSetup      string  `json:"weight_setup"`
}

type Exercise struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null" json:"name"`
	RepRollover uint `json:"rep_rollover"`
	WorkoutPlans []WorkoutPlan `gorm:"many2many:workout_plan_exercises;" json:"workout_plans"`
}

type Cardio struct {
    gorm.Model
    WorkoutLogID uint   `json:"workout_log_id" gorm:"uniqueIndex;not null"`
    Minutes      int    `json:"minutes"`
    Type         string `json:"type"`
}