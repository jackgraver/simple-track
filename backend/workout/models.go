package workout

import "time"

// NOTE: DB/storage is intentionally not implemented yet.
// These types model the intended domain for workouts.

type Exercise struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

type ExerciseSet struct {
    ID         uint    `json:"id"`
    ExerciseID uint    `json:"exercise_id"`
    Reps       int     `json:"reps"`
    Weight     float32 `json:"weight"`
}

type DayExercise struct {
    ID            uint           `json:"id"`
    WorkoutDayID  uint           `json:"workout_day_id"`
    ExerciseID    uint           `json:"exercise_id"`
    Exercise      Exercise       `json:"exercise"`
    Sets          []ExerciseSet  `json:"sets"`
}

type WorkoutDay struct {
    ID        uint          `json:"id"`
    Name      string        `json:"name"` // e.g. "Push", "Pull", "Legs"
    Date      time.Time     `json:"date"`
    Exercises []DayExercise `json:"exercises"`
}


