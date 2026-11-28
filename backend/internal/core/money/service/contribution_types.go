package service

import (
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"strings"
	"time"

	"be-simpletracker/internal/core/money/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Contribution types are a shared catalog (everyone reads, anyone adds, the creator or the admin
// edits, only the admin deletes); each user's own room figures live in models.ContributionRoom.
// docs/multi-user/DECISIONS.md, "C4 money".

// ContributionTypeInput is the shared rules of a type. It is the request body for editing a type.
type ContributionTypeInput struct {
	Name                   string   `json:"name"`
	EligibilityAge         *int     `json:"eligibility_age"`
	CarryForwardCap        *float64 `json:"carry_forward_cap"`
	LifetimeLimit          *float64 `json:"lifetime_limit"`
	WithdrawalsRestoreRoom bool     `json:"withdrawals_restore_room"`
}

// RoomInput is the caller's own figures for a type. All nil clears them.
type RoomInput struct {
	StartYear       *int     `json:"start_year"`
	OpeningRoomYear *int     `json:"opening_room_year"`
	OpeningRoom     *float64 `json:"opening_room"`
}

func (r RoomInput) empty() bool {
	return r.StartYear == nil && r.OpeningRoomYear == nil && r.OpeningRoom == nil
}

// ContributionTypeCreateInput creates a type with the caller's own room figures in one go.
type ContributionTypeCreateInput struct {
	ContributionTypeInput
	RoomInput
}

type AccountRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ContributionTypeSummary struct {
	models.ContributionType
	// The caller's own room figures for this type (null when they have none).
	StartYear       *int     `json:"start_year"`
	OpeningRoomYear *int     `json:"opening_room_year"`
	OpeningRoom     *float64 `json:"opening_room"`
	// CanEdit is true for the type's creator and the admin.
	CanEdit  bool         `json:"can_edit"`
	Accounts []AccountRef `json:"accounts"`
	// Room is nil when there is not enough configuration to calculate it
	// (no start year, eligibility or limits).
	Room *RoomSummary `json:"room"`
	// RoomIssue explains why Room is nil.
	RoomIssue string `json:"room_issue,omitempty"`
}

func canEditType(id database.Identity, contributionType models.ContributionType) bool {
	return id.IsAdmin || contributionType.CreatedBy == id.UserID
}

func ListContributionTypes(ctx context.Context, birthYear *int, now time.Time) ([]ContributionTypeSummary, error) {
	id, err := caller(ctx)
	if err != nil {
		return nil, err
	}
	db := database.From(ctx)
	var types []models.ContributionType
	if err := db.Preload("Limits", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("year DESC")
	}).Order("name ASC").Find(&types).Error; err != nil {
		return nil, err
	}
	var rooms []models.ContributionRoom
	if err := db.Scopes(database.Owned(ctx)).Find(&rooms).Error; err != nil {
		return nil, err
	}
	roomByType := make(map[uint]*models.ContributionRoom, len(rooms))
	for i := range rooms {
		roomByType[rooms[i].ContributionTypeID] = &rooms[i]
	}
	var accounts []models.Account
	if err := db.Scopes(database.Owned(ctx)).Where("investment_account_type_id IS NOT NULL").Order("name ASC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	flows, err := contributionFlows(ctx, db)
	if err != nil {
		return nil, err
	}
	summaries := make([]ContributionTypeSummary, 0, len(types))
	for _, contributionType := range types {
		room := roomByType[contributionType.ID]
		summary := ContributionTypeSummary{
			ContributionType: contributionType,
			CanEdit:          canEditType(id, contributionType),
			Accounts:         []AccountRef{},
		}
		if room != nil {
			summary.StartYear, summary.OpeningRoomYear, summary.OpeningRoom = room.StartYear, room.OpeningRoomYear, room.OpeningRoom
		}
		for _, account := range accounts {
			if account.ContributionTypeID != nil && *account.ContributionTypeID == contributionType.ID {
				summary.Accounts = append(summary.Accounts, AccountRef{ID: account.ID, Name: account.Name})
			}
		}
		summary.Room, summary.RoomIssue = roomFor(contributionType, room, birthYear, flows[contributionType.ID], now.Year())
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
	return nil, database.ErrNotFound
}

func roomFor(contributionType models.ContributionType, room *models.ContributionRoom, birthYear *int, flows map[int]YearFlow, currentYear int) (*RoomSummary, string) {
	startYear, ok := ResolveStartYear(contributionType, room, birthYear)
	if !ok {
		if contributionType.EligibilityAge != nil && birthYear == nil {
			return nil, "Set your birth year in settings, or a start year, to calculate room."
		}
		return nil, "Add a start year, opening room or an annual limit to calculate room."
	}
	if startYear > currentYear {
		return nil, "Room starts accruing in a future year."
	}
	rules := RoomRules{
		StartYear:              startYear,
		Limits:                 contributionType.Limits,
		CarryForwardCap:        contributionType.CarryForwardCap,
		LifetimeLimit:          contributionType.LifetimeLimit,
		WithdrawalsRestoreRoom: contributionType.WithdrawalsRestoreRoom,
	}
	if room != nil {
		rules.OpeningRoomYear, rules.OpeningRoom = room.OpeningRoomYear, room.OpeningRoom
	}
	summary := CalculateRoom(rules, flows, currentYear)
	return &summary, ""
}

// contributionFlows totals the caller's deposits and withdrawals per contribution type and year
// across their own (non-deleted) accounts assigned to that type. Deposits are scoped to the
// caller; the composite FK guarantees each deposit's account has the same owner.
func contributionFlows(ctx context.Context, db *gorm.DB) (map[uint]map[int]YearFlow, error) {
	type row struct {
		TypeID uint
		Kind   models.TransactionKind
		Amount float64
		Date   time.Time
	}
	var rows []row
	if err := db.Table("investment_deposits").Scopes(database.Owned(ctx)).
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

func CreateContributionType(ctx context.Context, input ContributionTypeCreateInput) (*models.ContributionType, error) {
	db := database.From(ctx)
	contributionType := models.ContributionType{}
	applyContributionTypeInput(&contributionType, input.ContributionTypeInput)
	if err := validateContributionType(db, contributionType); err != nil {
		return nil, err
	}
	if err := validateRoom(input.RoomInput); err != nil {
		return nil, err
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(&contributionType).Error; err != nil {
			return err
		}
		return saveRoom(ctx, tx, contributionType.ID, input.RoomInput)
	})
	if err != nil {
		return nil, err
	}
	return &contributionType, nil
}

// UpdateContributionType edits the shared rules of a type: its creator or the admin only.
func UpdateContributionType(ctx context.Context, id uint, input ContributionTypeInput) (*models.ContributionType, error) {
	identity, err := caller(ctx)
	if err != nil {
		return nil, err
	}
	db := database.From(ctx)
	var contributionType models.ContributionType
	if err := db.First(&contributionType, id).Error; err != nil {
		return nil, err
	}
	if !canEditType(identity, contributionType) {
		return nil, ErrForbidden
	}
	applyContributionTypeInput(&contributionType, input)
	if err := validateContributionType(db, contributionType); err != nil {
		return nil, err
	}
	if err := db.Omit(clause.Associations).Save(&contributionType).Error; err != nil {
		return nil, err
	}
	return &contributionType, nil
}

// SetContributionRoom replaces the caller's own room figures for a type (all nil clears them).
func SetContributionRoom(ctx context.Context, id uint, input RoomInput) error {
	db := database.From(ctx)
	if err := validateRoom(input); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var contributionType models.ContributionType
		if err := tx.Select("id").First(&contributionType, id).Error; err != nil {
			return err
		}
		return saveRoom(ctx, tx, id, input)
	})
}

// saveRoom upserts the caller's room row for a type; an empty input removes it.
func saveRoom(ctx context.Context, tx *gorm.DB, typeID uint, input RoomInput) error {
	var row models.ContributionRoom
	err := tx.Scopes(database.Owned(ctx)).Where("investment_account_type_id = ?", typeID).First(&row).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if input.empty() {
			return nil
		}
		row = models.ContributionRoom{
			ContributionTypeID: typeID,
			StartYear:          input.StartYear,
			OpeningRoomYear:    input.OpeningRoomYear,
			OpeningRoom:        input.OpeningRoom,
		}
		return tx.Omit(clause.Associations).Create(&row).Error
	case err != nil:
		return err
	case input.empty():
		return tx.Scopes(database.Owned(ctx)).Delete(&row).Error
	}
	row.StartYear, row.OpeningRoomYear, row.OpeningRoom = input.StartYear, input.OpeningRoomYear, input.OpeningRoom
	return tx.Omit(clause.Associations).Save(&row).Error
}

type PresetInput struct {
	Key             string   `json:"key"`
	Name            string   `json:"name"`
	StartYear       *int     `json:"start_year"`
	OpeningRoomYear *int     `json:"opening_room_year"`
	OpeningRoom     *float64 `json:"opening_room"`
}

// CreateContributionTypeFromPreset creates a contribution type with the preset's rules and limit
// schedule in one go, plus the caller's own room figures. The catalog is shared: when a type with
// that name already exists (another user added their TFSA first) it is reused as is and only the
// caller's room figures are saved.
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
	room := RoomInput{StartYear: input.StartYear, OpeningRoomYear: input.OpeningRoomYear, OpeningRoom: input.OpeningRoom}
	if err := validateRoom(room); err != nil {
		return nil, err
	}
	var existing models.ContributionType
	err := db.Where("LOWER(name) = LOWER(?)", name).First(&existing).Error
	switch {
	case err == nil:
		if err := saveRoom(ctx, db, existing.ID, room); err != nil {
			return nil, err
		}
		return &existing, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}
	contributionType := models.ContributionType{}
	applyContributionTypeInput(&contributionType, ContributionTypeInput{
		Name:                   name,
		EligibilityAge:         preset.EligibilityAge,
		CarryForwardCap:        preset.CarryForwardCap,
		LifetimeLimit:          preset.LifetimeLimit,
		WithdrawalsRestoreRoom: preset.WithdrawalsRestoreRoom,
	})
	if err := validateContributionType(db, contributionType); err != nil {
		return nil, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(&contributionType).Error; err != nil {
			return err
		}
		for _, limit := range preset.Limits {
			row := models.ContributionLimit{ContributionTypeID: contributionType.ID, Year: limit.Year, AnnualLimit: limit.AnnualLimit}
			if err := tx.Omit(clause.Associations).Create(&row).Error; err != nil {
				return err
			}
		}
		return saveRoom(ctx, tx, contributionType.ID, room)
	})
	if err != nil {
		return nil, err
	}
	return &contributionType, nil
}

// DeleteContributionType removes a type from the catalog: the admin only. It is refused while any
// account still uses it (the admin only sees their own accounts, so the database's foreign key
// answers for other users' accounts).
func DeleteContributionType(ctx context.Context, id uint) error {
	identity, err := caller(ctx)
	if err != nil {
		return err
	}
	if !identity.IsAdmin {
		return ErrForbidden
	}
	db := database.From(ctx)
	var count int64
	if err := db.Scopes(database.Owned(ctx)).Model(&models.Account{}).Where("investment_account_type_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return invalid("this contribution type is assigned to %d account(s); reassign them first", count)
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("investment_account_type_id = ?", id).Delete(&models.ContributionLimit{}).Error; err != nil {
			return err
		}
		return database.Affected(tx.Unscoped().Delete(&models.ContributionType{}, id))
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return invalid("this contribution type is still assigned to other users' accounts")
	}
	return err
}

// UpsertContributionLimit sets the limit for a year: the type's creator or the admin only.
func UpsertContributionLimit(ctx context.Context, contributionTypeID uint, year int, annualLimit float64) (*models.ContributionLimit, error) {
	identity, err := caller(ctx)
	if err != nil {
		return nil, err
	}
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
	if !canEditType(identity, contributionType) {
		return nil, ErrForbidden
	}
	var limit models.ContributionLimit
	err = db.Unscoped().Where("investment_account_type_id = ? AND year = ?", contributionTypeID, year).First(&limit).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		limit = models.ContributionLimit{ContributionTypeID: contributionTypeID, Year: year, AnnualLimit: annualLimit}
		if err := db.Omit(clause.Associations).Create(&limit).Error; err != nil {
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
	if err := db.Unscoped().Omit(clause.Associations).Save(&limit).Error; err != nil {
		return nil, err
	}
	return &limit, nil
}

// DeleteContributionLimit removes a year's limit: the admin only.
func DeleteContributionLimit(ctx context.Context, contributionTypeID uint, year int) error {
	identity, err := caller(ctx)
	if err != nil {
		return err
	}
	if !identity.IsAdmin {
		return ErrForbidden
	}
	db := database.From(ctx)
	return database.Affected(db.Unscoped().Where("investment_account_type_id = ? AND year = ?", contributionTypeID, year).Delete(&models.ContributionLimit{}))
}

func applyContributionTypeInput(contributionType *models.ContributionType, input ContributionTypeInput) {
	contributionType.Name = strings.TrimSpace(input.Name)
	contributionType.EligibilityAge = input.EligibilityAge
	contributionType.CarryForwardCap = input.CarryForwardCap
	contributionType.LifetimeLimit = input.LifetimeLimit
	contributionType.WithdrawalsRestoreRoom = input.WithdrawalsRestoreRoom
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
	if contributionType.EligibilityAge != nil && (*contributionType.EligibilityAge < 0 || *contributionType.EligibilityAge > 120) {
		return invalid("eligibility age must be between 0 and 120")
	}
	if contributionType.CarryForwardCap != nil && *contributionType.CarryForwardCap < 0 {
		return invalid("carry-forward cap must not be negative")
	}
	if contributionType.LifetimeLimit != nil && *contributionType.LifetimeLimit < 0 {
		return invalid("lifetime limit must not be negative")
	}
	return nil
}

func validateRoom(room RoomInput) error {
	if room.StartYear != nil && !validYear(*room.StartYear) {
		return invalid("start year must be between 1900 and 9999")
	}
	if (room.OpeningRoomYear == nil) != (room.OpeningRoom == nil) {
		return invalid("opening room needs both a year and an amount")
	}
	if room.OpeningRoomYear != nil && !validYear(*room.OpeningRoomYear) {
		return invalid("opening room year must be between 1900 and 9999")
	}
	return nil
}
