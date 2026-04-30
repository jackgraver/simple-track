package water

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

func CreateWaterLog(db *gorm.DB, date time.Time, amountOz float64, presetID *uint) (*WaterLog, error) {
	if amountOz <= 0 {
		return nil, errors.New("amount_oz must be positive")
	}
	if presetID != nil {
		var preset DrinkSizePreset
		if err := db.First(&preset, *presetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("preset not found")
			}
			return nil, err
		}
	}
	row := WaterLog{
		Date:     date,
		AmountOz: amountOz,
		PresetID: presetID,
	}
	if err := db.Create(&row).Error; err != nil {
		return nil, err
	}
	if err := db.Preload("Preset").First(&row, row.ID).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func ListWaterLogsForDate(db *gorm.DB, date time.Time) ([]WaterLog, error) {
	var rows []WaterLog
	err := db.Where("date = ?", date).Preload("Preset").Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func DeleteWaterLog(db *gorm.DB, id uint) error {
	res := db.Delete(&WaterLog{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ListDrinkSizePresets(db *gorm.DB) ([]DrinkSizePreset, error) {
	var rows []DrinkSizePreset
	err := db.Order("name ASC").Find(&rows).Error
	return rows, err
}

func CreateDrinkSizePreset(db *gorm.DB, name string, amountOz float64) (*DrinkSizePreset, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if amountOz <= 0 {
		return nil, errors.New("amount_oz must be positive")
	}
	p := DrinkSizePreset{Name: name, AmountOz: amountOz}
	if err := db.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func UpdateDrinkSizePreset(db *gorm.DB, id uint, name string, amountOz float64) (*DrinkSizePreset, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if amountOz <= 0 {
		return nil, errors.New("amount_oz must be positive")
	}
	var row DrinkSizePreset
	if err := db.First(&row, id).Error; err != nil {
		return nil, err
	}
	row.Name = name
	row.AmountOz = amountOz
	if err := db.Save(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func DeleteDrinkSizePreset(db *gorm.DB, id uint) error {
	res := db.Delete(&DrinkSizePreset{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
