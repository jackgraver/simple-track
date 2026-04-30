package water

import (
	"time"

	"gorm.io/gorm"
)

type DrinkSizePreset struct {
	gorm.Model
	Name     string  `json:"name" gorm:"not null"`
	AmountOz float64 `json:"amount_oz" gorm:"not null"`
}

func (DrinkSizePreset) TableName() string { return "drink_size_presets" }

type WaterLog struct {
	gorm.Model
	Date     time.Time        `json:"date" gorm:"index;not null"`
	AmountOz float64          `json:"amount_oz" gorm:"not null"`
	PresetID *uint            `json:"preset_id"`
	Preset   *DrinkSizePreset `json:"preset,omitempty" gorm:"foreignKey:PresetID"`
}

func (WaterLog) TableName() string { return "water_logs" }
