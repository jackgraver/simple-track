package workout

import "gorm.io/gorm"

// NOTE: No database calls yet. Keep functions simple and return TODO errors.

func GetToday(database *gorm.DB) (WorkoutDay, error) {
    return WorkoutDay{}, ErrTodo
}

func GetWeek(database *gorm.DB) ([]WorkoutDay, error) {
    return nil, ErrTodo
}

func GetAllExercises(database *gorm.DB) ([]Exercise, error) {
    return nil, ErrTodo
}

func AddExercise(database *gorm.DB, name string) (Exercise, error) {
    return Exercise{}, ErrTodo
}

// ErrTodo indicates unimplemented storage behavior.
var ErrTodo = gorm.ErrNotImplemented


