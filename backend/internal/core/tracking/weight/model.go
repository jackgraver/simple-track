package weight

import (
	"time"

	"gorm.io/gorm"
)

// BodyWeightLog is one weigh-in per day. WeightLbs is always set; the remaining
// body-composition metrics come from a Renpho import and are nil for manual entries.
type BodyWeightLog struct {
	gorm.Model
	Date      time.Time `json:"date" gorm:"uniqueIndex;not null"`
	WeightLbs float64   `json:"weight_lbs" gorm:"not null"`

	MeasuredAt *time.Time `json:"measured_at"`

	BMI                        *float64 `json:"bmi"`
	BodyFatPct                 *float64 `json:"body_fat_pct"`
	BodyFatMassLbs             *float64 `json:"body_fat_mass_lbs"`
	MusclePct                  *float64 `json:"muscle_pct"`
	MuscleMassLbs              *float64 `json:"muscle_mass_lbs"`
	SkeletalMusclePct          *float64 `json:"skeletal_muscle_pct"`
	SkeletalMuscleMassLbs      *float64 `json:"skeletal_muscle_mass_lbs"`
	BonePct                    *float64 `json:"bone_pct"`
	BoneMassLbs                *float64 `json:"bone_mass_lbs"`
	ProteinPct                 *float64 `json:"protein_pct"`
	ProteinMassLbs             *float64 `json:"protein_mass_lbs"`
	BodyWaterPct               *float64 `json:"body_water_pct"`
	BodyWaterMassLbs           *float64 `json:"body_water_mass_lbs"`
	FatFreeMassLbs             *float64 `json:"fat_free_mass_lbs"`
	SubcutaneousFatPct         *float64 `json:"subcutaneous_fat_pct"`
	VisceralFat                *float64 `json:"visceral_fat"`
	BMRKcal                    *float64 `json:"bmr_kcal" gorm:"column:bmr_kcal"`
	MetabolicAge               *float64 `json:"metabolic_age"`
	WHR                        *float64 `json:"whr" gorm:"column:whr"`
	OptimalWeightLbs           *float64 `json:"optimal_weight_lbs"`
	WeightLevel                *string  `json:"weight_level"`
	BodyType                   *string  `json:"body_type"`
	TargetOptimalWeightLbs     *float64 `json:"target_optimal_weight_lbs"`
	TargetOptimalMuscleMassLbs *float64 `json:"target_optimal_muscle_mass_lbs"`
	TargetOptimalFatMassLbs    *float64 `json:"target_optimal_fat_mass_lbs"`
	Remarks                    *string  `json:"remarks"`
}

func (BodyWeightLog) TableName() string { return "body_weight_logs" }
