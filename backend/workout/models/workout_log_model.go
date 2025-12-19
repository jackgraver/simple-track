package models

import (
	"be-simpletracker/database/repository"
	"time"

	"gorm.io/gorm"
)

// WorkoutLog represents a single workout session on a specific date
// Contains the exercises performed, sets logged, and optional cardio activity
type WorkoutLog struct {
	gorm.Model
	Date          time.Time `json:"date"`
	WorkoutPlanID *uint     `json:"workout_plan_id"`
	// WorkoutPlan is now in workout/plans package - use ID reference only
	// For preloading, use: Preload("WorkoutPlan") where WorkoutPlan is from workout/plans
	Exercises []LoggedExercise `json:"exercises" gorm:"constraint:OnDelete:CASCADE;"`
	Cardio    *Cardio          `json:"cardio" gorm:"constraint:OnDelete:CASCADE;"`
}

func (w WorkoutLog) GetID() uint        { return w.ID }
func (w WorkoutLog) TableName() string  { return "workout_logs" }
func (w WorkoutLog) GetDate() time.Time { return w.Date }
func (w WorkoutLog) Preloads() []string {
	// WorkoutPlan preload handled separately since it's in different package
	return []string{"Cardio", "Exercises.Sets", "Exercises.Exercise"}
}

// Verify interface implementations at compile time
var (
	_ repository.Entity     = (*WorkoutLog)(nil)
	_ repository.Preloadable = (*WorkoutLog)(nil)
)