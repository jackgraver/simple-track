package models

import (
	"be-simpletracker/database/repository"

	"gorm.io/gorm"
)

// WorkoutPlan represents a workout plan containing a collection of exercises
// that can be assigned to workout logs for tracking training sessions
type WorkoutPlan struct {
	gorm.Model
	Name      string  `json:"name"`
	DayOfWeek *int    `json:"day_of_week" gorm:"uniqueIndex:idx_day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday, null=unassigned
	Exercises []Exercise  `gorm:"many2many:workout_plan_exercises;" json:"exercises"`
}	

// GetID implements repository.Entity interface
func (w WorkoutPlan) GetID() uint {
	return w.ID
}

// TableName implements repository.Entity interface
func (w WorkoutPlan) TableName() string {
	return "workout_plans"
}

// Preloads implements repository.Preloadable interface
func (w WorkoutPlan) Preloads() []string {
	return []string{"Exercises"}
}

// Verify interface implementations at compile time
var (
	_ repository.Entity     = (*WorkoutPlan)(nil)
	_ repository.Preloadable = (*WorkoutPlan)(nil)
)
