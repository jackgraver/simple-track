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
	}
	pull_plan := WorkoutPlan{
		Name: "Pull",
	}
	legs_plan := WorkoutPlan{
		Name: "Legs",
	}
	upper_plan := WorkoutPlan{
		Name: "Upper",
	}
	lower_plan := WorkoutPlan{
		Name: "Lower",
	}
	active_rest_plan := WorkoutPlan{
		Name: "Active Rest",
	}
	rest_plan := WorkoutPlan{
		Name: "Rest",
	}
	
	m.db.Create(&push_plan)
	m.db.Create(&pull_plan)
	m.db.Create(&legs_plan)
	m.db.Create(&upper_plan)
	m.db.Create(&lower_plan)
	m.db.Create(&active_rest_plan)
	m.db.Create(&rest_plan)

	now := time.Now()

	push_log := WorkoutLog{
		Date: time.Date(2025, time.October, 21, 0, 0, 0, 0, now.Location()),
		WorkoutPlan: &push_plan,
		Exercises: []LoggedExercise{
			{
				ExerciseID: inclinePress.ID,
				Sets: []LoggedSet{
					{Reps: 9, Weight: 40},
					{Reps: 8, Weight: 40},
				},
			},
			{
				ExerciseID: chestFly.ID,
				Sets: []LoggedSet{
					{Reps: 9, Weight: 75},
					{Reps: 8, Weight: 75},
				},
			},
			{
				ExerciseID: dips.ID,
				Sets: []LoggedSet{
					{Reps: 8, Weight: -110},
				},
			},
			{
				ExerciseID: latRaise.ID,
				Sets: []LoggedSet{
					{Reps: 15, Weight: 10},
					{Reps: 14, Weight: 10},
				},
			},
			{
				ExerciseID: shoulderPress.ID,
				Sets: []LoggedSet{
					{Reps: 8, Weight: 75},
					{Reps: 7, Weight: 75},
				},
			},
			{
				ExerciseID: JMPress.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 105},
					{Reps: 6, Weight: 105},
				},
			},
			{
				ExerciseID: extensions.ID,
				Sets: []LoggedSet{
					{Reps: 6, Weight: 40},
					{Reps: 6, Weight: 40},
				},
			},
		},
	}
	pull_log := WorkoutLog{
		Date: time.Date(2025, time.October, 22, 0, 0, 0, 0, now.Location()),
		WorkoutPlan: &pull_plan,
		Exercises: []LoggedExercise{
			{
				ExerciseID: barbellRows.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 95},
					{Reps: 6, Weight: 95},
				},
			},	
			{
				ExerciseID: facePulls.ID,
				Sets: []LoggedSet{
					{Reps: 10, Weight: 40},
					{Reps: 10, Weight: 40},
				},
			},
			{
				ExerciseID: pulldowns.ID,
				Sets: []LoggedSet{
					{Reps: 6, Weight: 100},
					{Reps: 6, Weight: 100},
				},
			},		
			{
				ExerciseID: cableRows.ID,
				Sets: []LoggedSet{
					{Reps: 9, Weight: 60},
					{Reps: 9, Weight: 60},
				},
			},	
			{
				ExerciseID: inclineCurls.ID,
				Sets: []LoggedSet{
					{Reps: 6, Weight: 20},
					{Reps: 6, Weight: 20},
				},
			},			
			{
				ExerciseID: hammerCurls.ID,
				Sets: []LoggedSet{
					{Reps: 6, Weight: 22.5},
					{Reps: 6, Weight: 22.5},
				},
			},
		},
	}
	legs_log := WorkoutLog{
		Date: time.Date(2025, time.October, 23, 0, 0, 0, 0, now.Location()),
		WorkoutPlan: &legs_plan,
		Exercises: []LoggedExercise{
			{
				ExerciseID: outerThigh.ID,
				Sets: []LoggedSet{
					{Reps: 11, Weight: 80},
				},
			},
			{
				ExerciseID: innerThigh.ID,
				Sets: []LoggedSet{
					{Reps: 12, Weight: 70},
				},
			},
			{
				ExerciseID: legExtensions.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
				},
			},
			{
				ExerciseID: hamstringCurls.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 75},
					{Reps: 7, Weight: 75},
				},
			},
			{
				ExerciseID: squat.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 115},
					{Reps: 8, Weight: 115},
				},
			},
			{
				ExerciseID: deadlift.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 115},
					{Reps: 6, Weight: 115},
				},
			},
			{
				ExerciseID: calfRaise.ID,
				Sets: []LoggedSet{
					{Reps: 13, Weight: 90},
					{Reps: 13, Weight: 90},
				},
			},
			{
				ExerciseID: abCrunches.ID,
				Sets: []LoggedSet{
					{Reps: 8, Weight: 110},
					{Reps: 8, Weight: 110},
				},
			},
		},
	}
	upper_log := WorkoutLog{
		Date: time.Date(2025, time.October, 25, 0, 0, 0, 0, now.Location()),
		WorkoutPlan: &upper_plan,
		Exercises: []LoggedExercise{
			{
				ExerciseID: inclinePress.ID,
				Sets: []LoggedSet{
					{Reps: 5, Weight: 45},
					{Reps: 5, Weight: 45},
				},
			},
			{
				ExerciseID: chestFly.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
				},
			},
			{
				ExerciseID: dips.ID,
				Sets: []LoggedSet{
					{Reps: 9, Weight: -100},
				},
			},
			{
				ExerciseID: latRaise.ID,
				Sets: []LoggedSet{
					{Reps: 16, Weight: 10},
					{Reps: 14, Weight: 10},
				},
			},
			{
				ExerciseID: shoulderPress.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 6, Weight: 80},
				},
			},
			{
				ExerciseID: barbellRows.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 95},
					{Reps: 7, Weight: 95},
				},
			},	
			{
				ExerciseID: facePulls.ID,
				Sets: []LoggedSet{
					{Reps: 8, Weight: 40},
					{Reps: 8, Weight: 40},
				},
			},
			{
				ExerciseID: pulldowns.ID,
				Sets: []LoggedSet{
					{Reps: 6, Weight: 100},
					{Reps: 6, Weight: 100},
				},
			},
			{
				ExerciseID: JMPress.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 105},
					{Reps: 7, Weight: 105},
				},
			},
			{
				ExerciseID: extensions.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 6, Weight: 80},
				},
			},		
			{
				ExerciseID: inclineCurls.ID,
				Sets: []LoggedSet{
					{Reps: 9, Weight: 20},
					{Reps: 7, Weight: 20},
				},
			},			
			{
				ExerciseID: hammerCurls.ID,
				Sets: []LoggedSet{
					{Reps: 5, Weight: 22.5},
					{Reps: 6, Weight: 22.5},
				},
			},
		},
	}
	lower_log := WorkoutLog{
		Date: time.Date(2025, time.October, 26, 0, 0, 0, 0, now.Location()),
		WorkoutPlan: &lower_plan,
		Exercises: []LoggedExercise{
			{
				ExerciseID: outerThigh.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
				},
			},
			{
				ExerciseID: innerThigh.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
				},
			},
			{
				ExerciseID: legExtensions.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
				},
			},
			{
				ExerciseID: hamstringCurls.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
				},
			},
			{
				ExerciseID: squat.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
				},
				WeightSetup: "2 35lbs",
			},
			{
				ExerciseID: deadlift.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
				},
				WeightSetup: "35 x 2",
			},
			{
				ExerciseID: hipExtensions.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
				},
			},
			{
				ExerciseID: legPress.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
				},
				WeightSetup: "2x45",
			},
			{
				ExerciseID: calfRaise.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
				},
				WeightSetup: "2 45 plates",
			},
			{
				ExerciseID: abCrunches.ID,
				Sets: []LoggedSet{
					{Reps: 7, Weight: 80},
					{Reps: 7, Weight: 80},
				},
			},
		},
	}
	active_rest_log := WorkoutLog{
		Date: time.Date(2025, time.October, 20, 0, 0, 0, 0, now.Location()),
		WorkoutPlan: &active_rest_plan,
		Exercises: []LoggedExercise{
		},
	}
	rest_log := WorkoutLog{
		Date: time.Date(2025, time.October, 24, 0, 0, 0, 0, now.Location()),
		WorkoutPlan: &rest_plan,
		Exercises: []LoggedExercise{},
	}

	m.db.Create(&push_log)
	m.db.Create(&pull_log)
	m.db.Create(&legs_log)
	m.db.Create(&upper_log)
	m.db.Create(&lower_log)
	m.db.Create(&active_rest_log)
	m.db.Create(&rest_log)

	year := 2025
	start := time.Date(year, time.September, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, now.Location())

	skip_dates := []time.Time{
		time.Date(2025, time.October, 20, 0, 0, 0, 0, now.Location()),
		time.Date(2025, time.October, 21, 0, 0, 0, 0, now.Location()),
		time.Date(2025, time.October, 22, 0, 0, 0, 0, now.Location()),
		time.Date(2025, time.October, 23, 0, 0, 0, 0, now.Location()),
		time.Date(2025, time.October, 24, 0, 0, 0, 0, now.Location()),
		time.Date(2025, time.October, 25, 0, 0, 0, 0, now.Location()),
		time.Date(2025, time.October, 26, 0, 0, 0, 0, now.Location()),
	}
	skipMap := make(map[time.Time]struct{})
	for _, d := range skip_dates {
		skipMap[d] = struct{}{}
	}

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		if _, skip := skipMap[date]; skip {
			continue
		}
		wl := WorkoutLog{
			Date: date,
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
	WeightSetup  string  	  `json:"weight_setup"`
	PercentChange float32     `json:"percent_change" gorm:"-"`
}

type LoggedSet struct {
    gorm.Model
    LoggedExerciseID uint    `json:"logged_exercise_id"`
    Reps             uint     `json:"reps"`
    Weight           float32 `json:"weight"`
}

type Exercise struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null" json:"name"`
	RepRollover uint `json:"rep_rollover"`
}

type Cardio struct {
    gorm.Model
    WorkoutLogID uint   `json:"workout_log_id" gorm:"uniqueIndex;not null"`
    Minutes      int    `json:"minutes"`
    Type         string `json:"type"`
}