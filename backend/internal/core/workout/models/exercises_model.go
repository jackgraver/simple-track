package models

import (
	"time"

	"gorm.io/gorm"
)

// LoggedExercise represents an exercise that was performed during a workout session
// It links an exercise to a workout log and contains the sets performed
type LoggedExercise struct {
	gorm.Model
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
	LoggedExerciseID uint `json:"logged_exercise_id"`
	// ClientID is a client-generated uuid that makes POST /workout/sets idempotent.
	// NULL for sets created before migration 0004 or through the legacy full-replace endpoint.
	ClientID    *string `json:"client_id" gorm:"type:uuid;uniqueIndex:idx_logged_sets_client_id"`
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
	Name        string           `gorm:"uniqueIndex;not null" json:"name"`
	RepRollover uint             `json:"rep_rollover"`
	Cues        string           `json:"cues"`
	LoadType    ExerciseLoadType `gorm:"type:text;not null;default:plate_loaded_with_bar" json:"load_type"`
	// Unilateral exercises are logged one side at a time (see LoggedSet.Side).
	Unilateral bool `gorm:"not null;default:false" json:"unilateral"`
	// MachineSettings holds per-exercise machine positions (seat 4, pad 2, pulley top),
	// shown on the logging screen. Read-only for now: no endpoint writes it yet.
	MachineSettings []MachineSetting `gorm:"type:jsonb;serializer:json" json:"machine_settings"`
	WorkoutPlans    []WorkoutPlan    `gorm:"many2many:workout_plan_exercises;" json:"workout_plans"`
}

// MachineSetting is one labelled machine position, e.g. {"label": "Seat", "value": "4"}.
type MachineSetting struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (e Exercise) TableName() string { return "exercises" }

// Cardio represents cardiovascular exercise data for a workout session
// Stores duration and type of cardio activity performed
type Cardio struct {
	gorm.Model
	WorkoutLogID uint   `json:"workout_log_id" gorm:"uniqueIndex;not null"`
	Minutes      int    `json:"minutes"`
	Type         string `json:"type"`
	Notes        string `json:"notes"`
}

func (c Cardio) TableName() string { return "cardios" }
