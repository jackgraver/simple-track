package water

import (
	"time"

	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

// DrinkSizePreset is owned: each user has their own sizes (migration 0011). The reference
// conversion for owned tables (docs/multi-user/DECISIONS.md, "B3 owner kit").
type DrinkSizePreset struct {
	gorm.Model
	database.Owner
	Name     string  `json:"name" gorm:"not null"`
	AmountOz float64 `json:"amount_oz" gorm:"not null"`
}

func (DrinkSizePreset) TableName() string { return "drink_size_presets" }

// WaterLog is owned (migration 0013). preset_id points at one of the same user's presets (composite FK).
type WaterLog struct {
	gorm.Model
	database.Owner
	Date     time.Time `json:"date" gorm:"index;not null"`
	AmountOz float64   `json:"amount_oz" gorm:"not null"`
	PresetID *uint     `json:"preset_id"`
	// ClientID makes creates idempotent on retry (optimistic UI). Unique when set (unique per user, partial index in migration 0013).
	ClientID *string          `json:"client_id,omitempty" gorm:"type:uuid"`
	Preset   *DrinkSizePreset `json:"preset,omitempty" gorm:"foreignKey:PresetID"`
}

func (WaterLog) TableName() string { return "water_logs" }
