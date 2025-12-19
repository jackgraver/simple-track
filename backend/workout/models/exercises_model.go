package models

import (
	"be-simpletracker/database/repository"

	"gorm.io/gorm"
)

// LoggedExercise represents an exercise that was performed during a workout session
// It links an exercise to a workout log and contains the sets performed
type LoggedExercise struct {
    gorm.Model
    WorkoutLogID uint         `json:"workout_log_id"`
    ExerciseID   uint         `json:"exercise_id"`
	Exercise     *Exercise    `json:"exercise"`
    Sets         []LoggedSet  `json:"sets" gorm:"constraint:OnDelete:CASCADE;"`
	Notes        string  	  `json:"notes"`
	PercentChange float32     `json:"percent_change" gorm:"-"`
}

func (l LoggedExercise) GetID() uint       { return l.ID }
func (l LoggedExercise) TableName() string { return "logged_exercises" }
func (l LoggedExercise) Preloads() []string { return []string{"Exercise", "Sets"} }

// LoggedSet represents a single set of an exercise performed during a workout
// Contains reps, weight, and weight setup information for tracking progression
type LoggedSet struct {
    gorm.Model
    LoggedExerciseID uint    `json:"logged_exercise_id"`
    Reps             uint     `json:"reps"`
    Weight           float32 `json:"weight"`
    WeightSetup      string  `json:"weight_setup"`
}

func (l LoggedSet) GetID() uint       { return l.ID }
func (l LoggedSet) TableName() string { return "logged_sets" }

// Exercise represents a type of exercise that can be performed in workouts
// Can be associated with multiple workout plans via many-to-many relationship
type Exercise struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null" json:"name"`
	RepRollover uint `json:"rep_rollover"`
	WorkoutPlans []WorkoutPlan `gorm:"many2many:workout_plan_exercises;" json:"workout_plans"`
}

func (e Exercise) GetID() uint       { return e.ID }
func (e Exercise) TableName() string { return "exercises" }
func (e Exercise) Preloads() []string { return []string{} }

// Cardio represents cardiovascular exercise data for a workout session
// Stores duration and type of cardio activity performed
type Cardio struct {
    gorm.Model
    WorkoutLogID uint   `json:"workout_log_id" gorm:"uniqueIndex;not null"`
    Minutes      int    `json:"minutes"`
    Type         string `json:"type"`
}

func (c Cardio) GetID() uint       { return c.ID }
func (c Cardio) TableName() string { return "cardios" }

// Verify interface implementations at compile time
var (
	_ repository.Entity     = (*LoggedExercise)(nil)
	_ repository.Preloadable = (*LoggedExercise)(nil)
	_ repository.Entity     = (*LoggedSet)(nil)
	_ repository.Entity     = (*Exercise)(nil)
	_ repository.Entity     = (*Cardio)(nil)
)