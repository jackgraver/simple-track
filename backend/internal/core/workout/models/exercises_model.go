package models

import (
	"be-simpletracker/internal/database"
	"time"

	"gorm.io/gorm"
)

// LoggedExercise represents an exercise that was performed during a workout session
// It links an exercise to a workout log and contains the sets performed
type LoggedExercise struct {
	gorm.Model
	database.Owner
	// One live row per (workout_log_id, exercise_id): idx_logged_exercises_workout_exercise
	// (migration 0004). The tags mirror it for the SQLite module tests.
	WorkoutLogID  uint        `json:"workout_log_id" gorm:"uniqueIndex:idx_logged_exercises_workout_exercise,where:deleted_at IS NULL"`
	ExerciseID    uint        `json:"exercise_id" gorm:"uniqueIndex:idx_logged_exercises_workout_exercise,where:deleted_at IS NULL"`
	Exercise      *Exercise   `json:"exercise"`
	Sets          []LoggedSet `json:"sets" gorm:"constraint:OnDelete:CASCADE;"`
	Notes         string      `json:"notes"`
	PercentChange float32     `json:"percent_change" gorm:"-"`
	LogDate       time.Time   `json:"log_date,omitempty" gorm:"-"`
}

func (l LoggedExercise) TableName() string { return "logged_exercises" }

// LoggedSet represents a single set of an exercise performed during a workout
// Contains reps, weight, and weight setup information for tracking progression
type LoggedSet struct {
	gorm.Model
	database.Owner
	LoggedExerciseID uint `json:"logged_exercise_id"`
	// ClientID is a client-generated uuid that makes POST /workout/sets idempotent.
	// NULL for sets created before migration 0004 or through the legacy full-replace endpoint.
	// Unique per user (idx_logged_sets_client_id on (user_id, client_id), migration 0017).
	ClientID    *string `json:"client_id" gorm:"type:uuid"`
	Reps        uint    `json:"reps"`
	Weight      float32 `json:"weight"`
	WeightSetup string  `json:"weight_setup"`
	// Side is "left" or "right" for a set logged on one side of a unilateral exercise,
	// nil for a bilateral set. A unilateral set number is a pair of rows, one per side.
	Side *string `json:"side" gorm:"type:text"`
	// PairID ties the two sides of one unilateral set number together: the client_id of the side
	// logged first, on both rows. Stored rather than derived from log order so deleting or moving
	// one side doesn't re-pair every later set. Nil on bilateral sets.
	PairID *string `json:"pair_id" gorm:"type:uuid"`
}

func (l LoggedSet) TableName() string { return "logged_sets" }

const (
	SideLeft  = "left"
	SideRight = "right"
)

// NormalizeSide returns side as stored ("left", "right" or nil for bilateral) and whether it was valid.
// An empty string means bilateral.
func NormalizeSide(side *string) (*string, bool) {
	if side == nil || *side == "" {
		return nil, true
	}
	switch *side {
	case SideLeft, SideRight:
		s := *side
		return &s, true
	}
	return nil, false
}

type ExerciseLoadType string

const (
	ExerciseLoadTypePlateLoadedWithBar    ExerciseLoadType = "plate_loaded_with_bar"
	ExerciseLoadTypePlateLoadedWithoutBar ExerciseLoadType = "plate_loaded_without_bar"
	ExerciseLoadTypePlateLoadedTotal      ExerciseLoadType = "plate_loaded_total"
	ExerciseLoadTypeWeightStack           ExerciseLoadType = "weight_stack"
	ExerciseLoadTypeFreeWeights           ExerciseLoadType = "free_weights"
)

func NormalizeExerciseLoadType(loadType ExerciseLoadType) ExerciseLoadType {
	switch loadType {
	case ExerciseLoadTypePlateLoadedWithoutBar, ExerciseLoadTypePlateLoadedTotal, ExerciseLoadTypeWeightStack, ExerciseLoadTypeFreeWeights:
		return loadType
	default:
		return ExerciseLoadTypePlateLoadedWithBar
	}
}

// Exercise represents a type of exercise that can be performed in workouts
// Can be associated with multiple workout plans via many-to-many relationship
type Exercise struct {
	gorm.Model
	database.Creator
	Name        string           `gorm:"uniqueIndex;not null" json:"name"`
	RepRollover uint             `json:"rep_rollover"`
	Cues        string           `json:"cues"`
	LoadType    ExerciseLoadType `gorm:"type:text;not null;default:plate_loaded_with_bar" json:"load_type"`
	// Unilateral exercises are logged one side at a time (see LoggedSet.Side).
	Unilateral bool `gorm:"not null;default:false" json:"unilateral"`
	// MachineSettings are the signed-in user's own machine positions for this exercise (seat 4,
	// pad 2, pulley top), shown on the logging screen. They live in exercise_settings, not on the
	// exercise: repositories fill this field for the caller (see workoutrepo.FillMachineSettings).
	MachineSettings []MachineSetting `gorm:"-" json:"machine_settings"`
	// CanEdit is true when the caller may change the shared exercise (its creator or the admin).
	CanEdit      bool          `gorm:"-" json:"can_edit"`
	WorkoutPlans []WorkoutPlan `gorm:"many2many:workout_plan_exercises;" json:"workout_plans"`
}

// MachineSetting is one labelled machine position, e.g. {"label": "Seat", "value": "4"}.
type MachineSetting struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (e Exercise) TableName() string { return "exercises" }

// ExerciseSetting is one user's personal data for a shared exercise: today its machine settings.
// Unique per (user_id, exercise_id).
type ExerciseSetting struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	database.Owner
	ExerciseID      uint             `gorm:"not null"`
	MachineSettings []MachineSetting `gorm:"type:jsonb;serializer:json"`
}

func (ExerciseSetting) TableName() string { return "exercise_settings" }

// Cardio represents cardiovascular exercise data for a workout session
// Stores duration and type of cardio activity performed
type Cardio struct {
	gorm.Model
	database.Owner
	WorkoutLogID uint   `json:"workout_log_id" gorm:"uniqueIndex;not null"`
	Minutes      int    `json:"minutes"`
	Type         string `json:"type"`
	Notes        string `json:"notes"`
}

func (c Cardio) TableName() string { return "cardios" }
