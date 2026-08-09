package models

import (
	"be-simpletracker/internal/database/repository"

	"gorm.io/gorm"
)

// WorkoutPlan represents a workout plan containing a collection of exercises
// that can be assigned to workout logs for tracking training sessions
type WorkoutPlan struct {
	gorm.Model
	Name                 string          `json:"name"`
	WorkoutProgramID     *uint           `json:"workout_program_id"`
	WorkoutProgram       *WorkoutProgram `json:"workout_program,omitempty" gorm:"foreignKey:WorkoutProgramID"`
	DayOfWeek            *int            `json:"day_of_week,omitempty"`            // Legacy primary assignment; use AssignedDays for schedules.
	PlannedCardioType    string          `json:"planned_cardio_type,omitempty"`    // e.g. Run, Bike; empty means no planned cardio
	PlannedCardioMinutes int             `json:"planned_cardio_minutes,omitempty"` // Defaults the cardio log time for this plan.
	PreMobilityItems     []string        `json:"pre_mobility_items,omitempty" gorm:"type:jsonb;serializer:json"`
	PostMobilityItems    []string        `json:"post_mobility_items,omitempty" gorm:"type:jsonb;serializer:json"`
	Exercises            []Exercise      `gorm:"many2many:workout_plan_exercises;" json:"exercises"`
	AssignedDays         []int           `json:"assigned_days" gorm:"-"`
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
	_ repository.Entity      = (*WorkoutPlan)(nil)
	_ repository.Preloadable = (*WorkoutPlan)(nil)
)
