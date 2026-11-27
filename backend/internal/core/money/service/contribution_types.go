package service

import (
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"strings"
	"time"

	"be-simpletracker/internal/core/money/models"

	"gorm.io/gorm"
)

type ContributionTypeInput struct {
	Name                   string   `json:"name"`
	StartYear              *int     `json:"start_year"`
	EligibilityAge         *int     `json:"eligibility_age"`
	CarryForwardCap        *float64 `json:"carry_forward_cap"`
	LifetimeLimit          *float64 `json:"lifetime_limit"`
	WithdrawalsRestoreRoom bool     `json:"withdrawals_restore_room"`
	OpeningRoomYear        *int     `json:"opening_room_year"`
	OpeningRoom            *float64 `json:"opening_room"`
}

type AccountRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ContributionTypeSummary struct {
	models.ContributionType
	Accounts []AccountRef `json:"accounts"`
	// Room is nil when there is not enough configuration to calculate it
	// (no start year, eligibility or limits).
	Room *RoomSummary `json:"room"`
	// RoomIssue explains why Room is nil.
	RoomIssue string `json:"room_issue,omitempty"`
}

func ListContributionTypes(ctx context.Context, birthYear *int, now time.Time) ([]ContributionTypeSummary, error) {
	db := database.From(ctx)
	var types []models.ContributionType
	if err := db.Preload("Limits", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("year DESC")
	}).Order("name ASC").Find(&types).Error; err != nil {
		return nil, err
	}
	var accounts []models.Account
	if err := db.Where("investment_account_type_id IS NOT NULL").Order("name ASC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	flows, err := contributionFlows(db)
	if err != nil {
		return nil, err
	}
	summaries := make([]ContributionTypeSummary, 0, len(types))
	for _, contributionType := range types {
		summary := ContributionTypeSummary{ContributionType: contributionType, Accounts: []AccountRef{}}
		for _, account := range accounts {
			if account.ContributionTypeID != nil && *account.ContributionTypeID == contributionType.ID {
				summary.Accounts = append(summary.Accounts, AccountRef{ID: account.ID, Name: account.Name})
			}
		}
		summary.Room, summary.RoomIssue = roomFor(contributionType, birthYear, flows[contributionType.ID], now.Year())
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func GetContributionType(ctx context.Context, id uint, birthYear *int, now time.Time) (*ContributionTypeSummary, error) {
	summaries, err := ListContributionTypes(ctx, birthYear, now)
	if err != nil {
		return nil, err
	}
	for i := range summaries {
		if summaries[i].ID == id {
			return &summaries[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func roomFor(contributionType models.ContributionType, birthYear *int, flows map[int]YearFlow, currentYear int) (*RoomSummary, string) {
	startYear, ok := ResolveStartYear(contributionType, birthYear)
	if !ok {
		if contributionType.EligibilityAge != nil && birthYear == nil {
			return nil, "Set your birth year in settings, or a start year, to calculate room."
		}
		return nil, "Add a start year, opening room or an annual limit to calculate room."
	}
	if startYear > currentYear {
		return nil, "Room starts accruing in a future year."
	}
	summary := CalculateRoom(RoomRules{
		StartYear:              startYear,
		Limits:                 contributionType.Limits,
		CarryForwardCap:        contributionType.CarryForwardCap,
		LifetimeLimit:          contributionType.LifetimeLimit,
		WithdrawalsRestoreRoom: contributionType.WithdrawalsRestoreRoom,
		OpeningRoomYear:        contributionType.OpeningRoomYear,
		OpeningRoom:            contributionType.OpeningRoom,
	}, flows, currentYear)
	return &summary, ""
}

// contributionFlows totals deposits and withdrawals per contribution type and
// year across every (non-deleted) account assigned to that type.
func contributionFlows(db *gorm.DB) (map[uint]map[int]YearFlow, error) {
	type row struct {
		TypeID uint
		Kind   models.TransactionKind
		Amount float64
		Date   time.Time
	}
	var rows []row
	if err := db.Table("investment_deposits").
		Select("investment_accounts.investment_account_type_id AS type_id, investment_deposits.kind, investment_deposits.amount, investment_deposits.date").
		Joins("JOIN investment_accounts ON investment_accounts.id = investment_deposits.account_id").
		Where("investment_deposits.deleted_at IS NULL AND investment_accounts.deleted_at IS NULL AND investment_accounts.investment_account_type_id IS NOT NULL").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	flows := map[uint]map[int]YearFlow{}
	for _, r := range rows {
		byYear := flows[r.TypeID]
		if byYear == nil {
			byYear = map[int]YearFlow{}
			flows[r.TypeID] = byYear
		}
		year := r.Date.In(time.Local).Year()
		flow := byYear[year]
		if r.Kind == models.TransactionWithdrawal {
			flow.Withdrawn += r.Amount
		} else {
			flow.Contributed += r.Amount
		}
		byYear[year] = flow
	}
	return flows, nil
}

func CreateContributionType(ctx context.Context, input ContributionTypeInput) (*models.ContributionType, error) {
	db := database.From(ctx)
	contributionType := models.ContributionType{}
	applyContributionTypeInput(&contributionType, input)
	if err := validateContributionType(db, contributionType); err != nil {
		return nil, err
	}
	if err := db.Create(&contributionType).Error; err != nil {
		return nil, err
	}
	return &contributionType, nil
}

func UpdateContributionType(ctx context.Context, id uint, input ContributionTypeInput) (*models.ContributionType, error) {
	db := database.From(ctx)
	var contributionType models.ContributionType
	if err := db.First(&contributionType, id).Error; err != nil {
		return nil, err
	}
	applyContributionTypeInput(&contributionType, input)
	if err := validateContributionType(db, contributionType); err != nil {
		return nil, err
	}
	if err := db.Save(&contributionType).Error; err != nil {
		return nil, err
	}
	return &contributionType, nil
}

type PresetInput struct {
	Key             string   `json:"key"`
	Name            string   `json:"name"`
	StartYear       *int     `json:"start_year"`
	OpeningRoomYear *int     `json:"opening_room_year"`
	OpeningRoom     *float64 `json:"opening_room"`
}

// CreateContributionTypeFromPreset creates a contribution type with the
// preset's rules and limit schedule in one go.
func CreateContributionTypeFromPreset(ctx context.Context, input PresetInput) (*models.ContributionType, error) {
	db := database.From(ctx)
	preset, ok := findContributionPreset(input.Key)
	if !ok {
		return nil, invalid("unknown preset %q", input.Key)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = preset.Name
	}
	contributionType := models.ContributionType{}
	applyContributionTypeInput(&contributionType, ContributionTypeInput{
		Name:                   name,
		StartYear:              input.StartYear,
		EligibilityAge:         preset.EligibilityAge,
		CarryForwardCap:        preset.CarryForwardCap,
		LifetimeLimit:          preset.LifetimeLimit,
		WithdrawalsRestoreRoom: preset.WithdrawalsRestoreRoom,
		OpeningRoomYear:        input.OpeningRoomYear,
		OpeningRoom:            input.OpeningRoom,
	})
	if err := validateContributionType(db, contributionType); err != nil {
		return nil, err
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&contributionType).Error; err != nil {
			return err
		}
		for _, limit := range preset.Limits {
			row := models.ContributionLimit{ContributionTypeID: contributionType.ID, Year: limit.Year, AnnualLimit: limit.AnnualLimit}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &contributionType, nil
}

func DeleteContributionType(ctx context.Context, id uint) error {
	db := database.From(ctx)
	var count int64
	if err := db.Model(&models.Account{}).Where("investment_account_type_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return invalid("this contribution type is assigned to %d account(s); reassign them first", count)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("investment_account_type_id = ?", id).Delete(&models.ContributionLimit{}).Error; err != nil {
			return err
		}
		result := tx.Unscoped().Delete(&models.ContributionType{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func UpsertContributionLimit(ctx context.Context, contributionTypeID uint, year int, annualLimit float64) (*models.ContributionLimit, error) {
	db := database.From(ctx)
	if !validYear(year) {
		return nil, invalid("year must be between 1900 and 9999")
	}
	if annualLimit < 0 {
		return nil, invalid("annual limit must not be negative")
	}
	var contributionType models.ContributionType
	if err := db.First(&contributionType, contributionTypeID).Error; err != nil {
		return nil, err
	}
	var limit models.ContributionLimit
	err := db.Unscoped().Where("investment_account_type_id = ? AND year = ?", contributionTypeID, year).First(&limit).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		limit = models.ContributionLimit{ContributionTypeID: contributionTypeID, Year: year, AnnualLimit: annualLimit}
		if err := db.Create(&limit).Error; err != nil {
			return nil, err
		}
		return &limit, nil
	}
	if err != nil {
		return nil, err
	}
	// Revive a soft-deleted row so the unique index stays satisfied.
	limit.DeletedAt = gorm.DeletedAt{}
	limit.AnnualLimit = annualLimit
	if err := db.Unscoped().Save(&limit).Error; err != nil {
		return nil, err
	}
	return &limit, nil
}

func DeleteContributionLimit(ctx context.Context, contributionTypeID uint, year int) error {
	db := database.From(ctx)
	result := db.Unscoped().Where("investment_account_type_id = ? AND year = ?", contributionTypeID, year).Delete(&models.ContributionLimit{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func applyContributionTypeInput(contributionType *models.ContributionType, input ContributionTypeInput) {
	contributionType.Name = strings.TrimSpace(input.Name)
	contributionType.StartYear = input.StartYear
	contributionType.EligibilityAge = input.EligibilityAge
	contributionType.CarryForwardCap = input.CarryForwardCap
	contributionType.LifetimeLimit = input.LifetimeLimit
	contributionType.WithdrawalsRestoreRoom = input.WithdrawalsRestoreRoom
	contributionType.OpeningRoomYear = input.OpeningRoomYear
	contributionType.OpeningRoom = input.OpeningRoom
}

func validateContributionType(db *gorm.DB, contributionType models.ContributionType) error {
	if contributionType.Name == "" {
		return invalid("name is required")
	}
	var existing int64
	if err := db.Model(&models.ContributionType{}).
		Where("LOWER(name) = LOWER(?) AND id <> ?", contributionType.Name, contributionType.ID).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return invalid("a contribution type named %q already exists", contributionType.Name)
	}
	if contributionType.StartYear != nil && !validYear(*contributionType.StartYear) {
		return invalid("start year must be between 1900 and 9999")
	}
	if contributionType.EligibilityAge != nil && (*contributionType.EligibilityAge < 0 || *contributionType.EligibilityAge > 120) {
		return invalid("eligibility age must be between 0 and 120")
	}
	if contributionType.CarryForwardCap != nil && *contributionType.CarryForwardCap < 0 {
		return invalid("carry-forward cap must not be negative")
	}
	if contributionType.LifetimeLimit != nil && *contributionType.LifetimeLimit < 0 {
		return invalid("lifetime limit must not be negative")
	}
	if (contributionType.OpeningRoomYear == nil) != (contributionType.OpeningRoom == nil) {
		return invalid("opening room needs both a year and an amount")
	}
	if contributionType.OpeningRoomYear != nil && !validYear(*contributionType.OpeningRoomYear) {
		return invalid("opening room year must be between 1900 and 9999")
	}
	return nil
}
