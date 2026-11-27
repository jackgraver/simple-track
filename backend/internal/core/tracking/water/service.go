package water

import (
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// CreateWaterLog inserts a log. With a non-empty clientID it is idempotent: a replay returns the
// existing row (even if it was deleted since) and creates nothing.
func CreateWaterLog(ctx context.Context, date time.Time, amountOz float64, presetID *uint, clientID string) (*WaterLog, error) {
	db := database.From(ctx)
	if clientID != "" {
		if row, err := findByClientID(db, clientID); err != nil || row != nil {
			return row, err
		}
	}
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
	if clientID != "" {
		row.ClientID = &clientID
	}
	if err := database.CreateInSavepoint(db, &row); err != nil {
		// Lost a race with a concurrent replay of the same client_id: return the winner's row.
		if clientID != "" {
			if existing, ferr := findByClientID(db, clientID); ferr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}
	if err := db.Preload("Preset").First(&row, row.ID).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func findByClientID(db *gorm.DB, clientID string) (*WaterLog, error) {
	var row WaterLog
	err := db.Unscoped().Preload("Preset").Where("client_id = ?", clientID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func ListWaterLogsForDate(ctx context.Context, date time.Time) ([]WaterLog, error) {
	db := database.From(ctx)
	var rows []WaterLog
	err := db.Where("date = ?", date).Preload("Preset").Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func DeleteWaterLog(ctx context.Context, id uint) error {
	db := database.From(ctx)
	res := db.Delete(&WaterLog{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteWaterLogByClientID deletes the entry created with clientID (gorm.ErrRecordNotFound if none).
func DeleteWaterLogByClientID(ctx context.Context, clientID string) error {
	db := database.From(ctx)
	res := db.Where("client_id = ?", clientID).Delete(&WaterLog{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ListDrinkSizePresets(ctx context.Context) ([]DrinkSizePreset, error) {
	db := database.From(ctx)
	var rows []DrinkSizePreset
	err := db.Order("name ASC").Find(&rows).Error
	return rows, err
}

func CreateDrinkSizePreset(ctx context.Context, name string, amountOz float64) (*DrinkSizePreset, error) {
	db := database.From(ctx)
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

func UpdateDrinkSizePreset(ctx context.Context, id uint, name string, amountOz float64) (*DrinkSizePreset, error) {
	db := database.From(ctx)
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

func DeleteDrinkSizePreset(ctx context.Context, id uint) error {
	db := database.From(ctx)
	res := db.Delete(&DrinkSizePreset{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
