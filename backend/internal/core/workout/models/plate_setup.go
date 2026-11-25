package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type PlateMode string

const (
	PlateModePerSide PlateMode = "per_side"
	PlateModeTotal   PlateMode = "total"
)

// PlateSetup is a snapshot of how a set was loaded. Plates maps a plate's weight ("45", "2.5")
// to how many are loaded: per side in per_side mode, in all in total mode. Base is the exercise's
// base weight at the time, so the set still reads right after the exercise's bar is changed.
type PlateSetup struct {
	Base   float32        `json:"base"`
	Mode   PlateMode      `json:"mode"`
	Plates map[string]int `json:"plates"`
}

const (
	maxPlateLbs      = 100
	maxPlateCount    = 50
	maxBaseWeightLbs = 2000
	maxWeightStepLbs = 200
)

// Validate rejects setups no client should send: unknown modes, negative or absurd weights and
// counts, and plate keys that aren't positive numbers.
func (p PlateSetup) Validate() error {
	if p.Mode != PlateModePerSide && p.Mode != PlateModeTotal {
		return fmt.Errorf("plate_setup.mode must be %q or %q", PlateModePerSide, PlateModeTotal)
	}
	if p.Base < 0 || p.Base > maxBaseWeightLbs {
		return fmt.Errorf("plate_setup.base must be between 0 and %d", maxBaseWeightLbs)
	}
	for key, count := range p.Plates {
		lbs, err := strconv.ParseFloat(key, 64)
		if err != nil || lbs <= 0 || lbs > maxPlateLbs {
			return fmt.Errorf("plate_setup.plates: %q is not a plate weight", key)
		}
		if count < 0 || count > maxPlateCount {
			return fmt.Errorf("plate_setup.plates[%s] must be between 0 and %d", key, maxPlateCount)
		}
	}
	return nil
}

// ValidateExerciseWeights checks an exercise's base weight and weight step (either may be nil).
func ValidateExerciseWeights(baseWeight, weightStep *float32) error {
	if baseWeight != nil && (*baseWeight < 0 || *baseWeight > maxBaseWeightLbs) {
		return fmt.Errorf("base_weight must be between 0 and %d", maxBaseWeightLbs)
	}
	if weightStep != nil && (*weightStep <= 0 || *weightStep > maxWeightStepLbs) {
		return fmt.Errorf("weight_step must be greater than 0 and at most %d", maxWeightStepLbs)
	}
	return nil
}

// Optional is a JSON field that tells "absent" (Set false) from "null" (Set true, Value nil), for
// requests where null clears a value and a missing field leaves it alone.
type Optional[T any] struct {
	Set   bool
	Value *T
}

func Some[T any](v T) Optional[T] { return Optional[T]{Set: true, Value: &v} }

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}
