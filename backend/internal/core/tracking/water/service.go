package water

import (
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Errors from CreateWaterLog that the client caused (400, message shown as is).
var (
	ErrPresetNotFound = errors.New("preset not found")
)

// CreateWaterLog inserts a log. With a non-empty clientID it is idempotent: a replay returns the
// existing row (even if it was deleted since) and creates nothing.
func CreateWaterLog(ctx context.Context, date time.Time, amountOz float64, presetID *uint, clientID string) (*WaterLog, error) {
	db := database.From(ctx)
	if clientID != "" {
		if row, err := findByClientID(ctx, db, clientID); err != nil || row != nil {
			return row, err
		}
	}
	if amountOz <= 0 {
		return nil, ErrAmountNotPositive
	}
	if presetID != nil {
		// Only the user's own presets (someone else's reads as missing).
		var preset DrinkSizePreset
		if err := db.Scopes(database.Owned(ctx)).First(&preset, *presetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPresetNotFound
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
			if existing, ferr := findByClientID(ctx, db, clientID); ferr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}
	if err := db.Preload("Preset", database.Owned(ctx)).First(&row, row.ID).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func findByClientID(ctx context.Context, db *gorm.DB, clientID string) (*WaterLog, error) {
	var row WaterLog
	err := db.Unscoped().Preload("Preset", database.Owned(ctx)).Where("client_id = ?", clientID).First(&row).Error
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
	err := db.Where("date = ?", date).Preload("Preset", database.Owned(ctx)).Order("created_at DESC").Find(&rows).Error
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

// Errors from the preset functions that the client caused (400, message shown as is).
var (
	ErrPresetNameRequired = errors.New("name is required")
	ErrAmountNotPositive  = errors.New("amount_oz must be positive")
)

// Drink size presets are owned (migration 0011): every query is scoped to the current user with
// database.Owned, creates are stamped by database.Owner, and someone else's preset is
// database.ErrNotFound, exactly like a missing one.

func ListDrinkSizePresets(ctx context.Context) ([]DrinkSizePreset, error) {
	var rows []DrinkSizePreset
	err := database.From(ctx).Scopes(database.Owned(ctx)).Order("name ASC").Find(&rows).Error
	return rows, err
}

func validatePreset(name string, amountOz float64) error {
	if name == "" {
		return ErrPresetNameRequired
	}
	if amountOz <= 0 {
		return ErrAmountNotPositive
	}
	return nil
}

func CreateDrinkSizePreset(ctx context.Context, name string, amountOz float64) (*DrinkSizePreset, error) {
	if err := validatePreset(name, amountOz); err != nil {
		return nil, err
	}
	p := DrinkSizePreset{Name: name, AmountOz: amountOz}
	if err := database.From(ctx).Omit(clause.Associations).Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func UpdateDrinkSizePreset(ctx context.Context, id uint, name string, amountOz float64) (*DrinkSizePreset, error) {
	if err := validatePreset(name, amountOz); err != nil {
		return nil, err
	}
	db := database.From(ctx)
	var row DrinkSizePreset
	if err := db.Scopes(database.Owned(ctx)).First(&row, id).Error; err != nil {
		return nil, err
	}
	res := db.Model(&row).Scopes(database.Owned(ctx)).Select("name", "amount_oz", "updated_at").
		Updates(DrinkSizePreset{Name: name, AmountOz: amountOz})
	if err := database.Affected(res); err != nil {
		return nil, err
	}
	row.Name, row.AmountOz = name, amountOz
	return &row, nil
}

func DeleteDrinkSizePreset(ctx context.Context, id uint) error {
	return database.Affected(database.From(ctx).Scopes(database.Owned(ctx)).Delete(&DrinkSizePreset{}, id))
}
