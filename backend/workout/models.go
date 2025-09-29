package workout

import (
	"time"

	"gorm.io/gorm"
)
func MigrateWorkoutDatabase(db *gorm.DB) {
	// db.Migrator().DropTable(&Food{}, &Meal{}, &MealItem{}, &DayGoals{}, &MealPlanDay{}, &DayMeal{})
	db.AutoMigrate(&Exercise{}, &ExerciseSet{}, &DayExercise{}, &WorkoutDay{}, &Cardio{})
	// seed(db)
}

type Cardio struct {
    Minutes int `json:"minutes"`
    Type    string `json:"type"`
}

type Exercise struct {
    gorm.Model
    Name string `json:"name"`
}

type ExerciseSet struct {
    gorm.Model
    ExerciseID uint    `json:"exercise_id"`
    Reps       int     `json:"reps"`
    Weight     float32 `json:"weight"`
}

type DayExercise struct {
    gorm.Model
    WorkoutDayID  uint           `json:"workout_day_id"`
    ExerciseID    uint           `json:"exercise_id"`
    Exercise      Exercise       `json:"exercise"`
    Sets          []ExerciseSet  `json:"sets"`
}

type WorkoutDay struct {
    gorm.Model
    Name      string        `json:"name"` // e.g. "Push", "Pull", "Legs"
    Date      time.Time     `json:"date"`
    Exercises []DayExercise `json:"exercises"`
    Cardio    Cardio        `json:"cardio"`
}