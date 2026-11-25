package models

import (
	"be-simpletracker/internal/database/repository"
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

func (l LoggedExercise) GetID() uint        { return l.ID }
func (l LoggedExercise) TableName() string  { return "logged_exercises" }
func (l LoggedExercise) Preloads() []string { return []string{"Exercise", "Sets"} }

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
	// PlateSetup is the plates the set was lifted with (plate-loaded exercises only). Weight stays
	// the logged total and can differ from what the plates add up to (a typed override). NULL for
	// sets logged before migration 0008 and for other load types; clients then fall back to
	// parsing WeightSetup.
	PlateSetup *PlateSetup `json:"plate_setup" gorm:"type:jsonb;serializer:json"`
}

func (l LoggedSet) GetID() uint       { return l.ID }
func (l LoggedSet) TableName() string { return "logged_sets" }

type ExerciseLoadType string

const (
	// ExerciseLoadTypePlateLoadedPerSide: plates on each side, plus the exercise's base weight
	// (bar or sled; 0 when it isn't counted).
	ExerciseLoadTypePlateLoadedPerSide ExerciseLoadType = "plate_loaded_per_side"
	// ExerciseLoadTypePlateLoadedTotal: every plate counted once (single-peg machines).
	ExerciseLoadTypePlateLoadedTotal ExerciseLoadType = "plate_loaded_total"
	ExerciseLoadTypeWeightStack      ExerciseLoadType = "weight_stack"
	ExerciseLoadTypeFreeWeights      ExerciseLoadType = "free_weights"

	// Legacy values from before migration 0008, still accepted from old clients. Both become
	// plate_loaded_per_side with base weight 45 / 0 (see LegacyLoadTypeBaseWeight).
	ExerciseLoadTypePlateLoadedWithBar    ExerciseLoadType = "plate_loaded_with_bar"
	ExerciseLoadTypePlateLoadedWithoutBar ExerciseLoadType = "plate_loaded_without_bar"
)

func NormalizeExerciseLoadType(loadType ExerciseLoadType) ExerciseLoadType {
	switch loadType {
	case ExerciseLoadTypePlateLoadedTotal, ExerciseLoadTypeWeightStack, ExerciseLoadTypeFreeWeights:
		return loadType
	default:
		return ExerciseLoadTypePlateLoadedPerSide
	}
}

// LegacyLoadTypeBaseWeight is the base weight a legacy load type implies (45 with the bar, 0
// without), or nil for any other value.
func LegacyLoadTypeBaseWeight(loadType ExerciseLoadType) *float32 {
	var base float32
	switch loadType {
	case ExerciseLoadTypePlateLoadedWithBar:
		base = 45
	case ExerciseLoadTypePlateLoadedWithoutBar:
		base = 0
	default:
		return nil
	}
	return &base
}

// Exercise represents a type of exercise that can be performed in workouts
// Can be associated with multiple workout plans via many-to-many relationship
type Exercise struct {
	gorm.Model
	Name        string           `gorm:"uniqueIndex;not null" json:"name"`
	RepRollover uint             `json:"rep_rollover"`
	Cues        string           `json:"cues"`
	LoadType    ExerciseLoadType `gorm:"type:text;not null;default:plate_loaded_per_side" json:"load_type"`
	// BaseWeight is what's loaded before any plates: the bar, sled or machine's starting weight.
	// nil means the load type's default (45 for plates per side, 0 otherwise).
	BaseWeight *float32 `json:"base_weight"`
	// WeightStep is the automatic increase (after every set hits the rollover) and the +/- step.
	// nil means the client's default, 5 lb.
	WeightStep *float32 `json:"weight_step"`
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

func (e Exercise) GetID() uint        { return e.ID }
func (e Exercise) TableName() string  { return "exercises" }
func (e Exercise) Preloads() []string { return []string{} }

// Cardio represents cardiovascular exercise data for a workout session
// Stores duration and type of cardio activity performed
type Cardio struct {
	gorm.Model
	WorkoutLogID uint   `json:"workout_log_id" gorm:"uniqueIndex;not null"`
	Minutes      int    `json:"minutes"`
	Type         string `json:"type"`
	Notes        string `json:"notes"`
}

func (c Cardio) GetID() uint       { return c.ID }
func (c Cardio) TableName() string { return "cardios" }

// Verify interface implementations at compile time
var (
	_ repository.Entity      = (*LoggedExercise)(nil)
	_ repository.Preloadable = (*LoggedExercise)(nil)
	_ repository.Entity      = (*LoggedSet)(nil)
	_ repository.Entity      = (*Exercise)(nil)
	_ repository.Entity      = (*Cardio)(nil)
)
