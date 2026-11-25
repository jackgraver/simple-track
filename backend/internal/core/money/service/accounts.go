package service

import (
	"be-simpletracker/internal/clock"
	"errors"
	"strings"
	"time"

	"be-simpletracker/internal/core/money/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountInput struct {
	Name               string             `json:"name"`
	Kind               models.AccountKind `json:"kind"`
	Institution        string             `json:"institution"`
	Notes              string             `json:"notes"`
	ContributionTypeID *uint              `json:"contribution_type_id"`
	Archived           bool               `json:"archived"`
}

type AccountSummary struct {
	models.Account
	IsLiability bool       `json:"is_liability"`
	BalanceDate *time.Time `json:"balance_date"`
	// PreviousBalance is the snapshot before the latest one, for showing
	// the most recent change.
	PreviousBalance     *float64   `json:"previous_balance"`
	PreviousBalanceDate *time.Time `json:"previous_balance_date"`
	TotalDeposits       float64    `json:"total_deposits"`
	TotalWithdrawals    float64    `json:"total_withdrawals"`
	NetContributions    float64    `json:"net_contributions"`
	TransactionCount    int64      `json:"transaction_count"`
	// Gain is balance minus net contributions. Only meaningful when
	// deposits have been recorded.
	Gain float64 `json:"gain"`
}

func ListAccounts(db *gorm.DB, includeArchived bool) ([]AccountSummary, error) {
	query := db.Preload("ContributionType").Order("name ASC")
	if !includeArchived {
		query = query.Where("archived_at IS NULL")
	}
	var accounts []models.Account
	if err := query.Find(&accounts).Error; err != nil {
		return nil, err
	}
	return summarizeAccounts(db, accounts)
}

func GetAccount(db *gorm.DB, id uint) (*AccountSummary, error) {
	var account models.Account
	if err := db.Preload("ContributionType").First(&account, id).Error; err != nil {
		return nil, err
	}
	summaries, err := summarizeAccounts(db, []models.Account{account})
	if err != nil {
		return nil, err
	}
	return &summaries[0], nil
}

func summarizeAccounts(db *gorm.DB, accounts []models.Account) ([]AccountSummary, error) {
	summaries := make([]AccountSummary, 0, len(accounts))
	if len(accounts) == 0 {
		return summaries, nil
	}
	ids := make([]uint, len(accounts))
	for i, account := range accounts {
		ids[i] = account.ID
	}

	type totalRow struct {
		AccountID uint
		Kind      models.TransactionKind
		Total     float64
		Count     int64
	}
	var totals []totalRow
	if err := db.Model(&models.AccountTransaction{}).
		Select("account_id, kind, COALESCE(SUM(amount), 0) AS total, COUNT(*) AS count").
		Where("account_id IN ?", ids).
		Group("account_id, kind").
		Scan(&totals).Error; err != nil {
		return nil, err
	}

	var balances []models.AccountBalance
	if err := db.Where("account_id IN ?", ids).Order("account_id ASC, date DESC").Find(&balances).Error; err != nil {
		return nil, err
	}
	latest := map[uint][]models.AccountBalance{}
	for _, balance := range balances {
		if len(latest[balance.AccountID]) < 2 {
			latest[balance.AccountID] = append(latest[balance.AccountID], balance)
		}
	}

	for _, account := range accounts {
		summary := AccountSummary{Account: account, IsLiability: account.Kind.IsLiability()}
		for _, total := range totals {
			if total.AccountID != account.ID {
				continue
			}
			summary.TransactionCount += total.Count
			if total.Kind == models.TransactionWithdrawal {
				summary.TotalWithdrawals += total.Total
			} else {
				summary.TotalDeposits += total.Total
			}
		}
		snapshots := latest[account.ID]
		if len(snapshots) > 0 {
			date := snapshots[0].Date
			summary.BalanceDate = &date
			summary.CurrentBalance = snapshots[0].Balance
		}
		if len(snapshots) > 1 {
			previous := snapshots[1].Balance
			previousDate := snapshots[1].Date
			summary.PreviousBalance = &previous
			summary.PreviousBalanceDate = &previousDate
		}
		summary.NetContributions = roundCents(summary.TotalDeposits - summary.TotalWithdrawals)
		summary.Gain = roundCents(summary.CurrentBalance - summary.NetContributions)
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

// CreateAccount creates an account and, when given, records its opening
// balance as the first snapshot.
func CreateAccount(db *gorm.DB, input AccountInput, openingBalance *float64, balanceDate time.Time) (*models.Account, error) {
	account := models.Account{}
	applyAccountInput(&account, input, clock.Now())
	if err := validateAccount(db, account); err != nil {
		return nil, err
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		if openingBalance != nil {
			return setBalance(tx, account.ID, balanceDate, *openingBalance)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func UpdateAccount(db *gorm.DB, id uint, input AccountInput) (*models.Account, error) {
	var account models.Account
	if err := db.First(&account, id).Error; err != nil {
		return nil, err
	}
	applyAccountInput(&account, input, clock.Now())
	if err := validateAccount(db, account); err != nil {
		return nil, err
	}
	// Omit the association so a stale preloaded type is never written back.
	if err := db.Omit("ContributionType").Save(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// DeleteAccount permanently removes an account with its balances and
// transactions. Archive it instead to keep its history.
func DeleteAccount(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("account_id = ?", id).Delete(&models.AccountBalance{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("account_id = ?", id).Delete(&models.AccountTransaction{}).Error; err != nil {
			return err
		}
		result := tx.Unscoped().Delete(&models.Account{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func applyAccountInput(account *models.Account, input AccountInput, now time.Time) {
	account.Name = strings.TrimSpace(input.Name)
	account.Kind = input.Kind
	if account.Kind == "" {
		account.Kind = models.AccountKindOther
	}
	account.Institution = strings.TrimSpace(input.Institution)
	account.Notes = strings.TrimSpace(input.Notes)
	account.ContributionTypeID = input.ContributionTypeID
	account.ContributionType = nil
	switch {
	case input.Archived && account.ArchivedAt == nil:
		account.ArchivedAt = &now
	case !input.Archived:
		account.ArchivedAt = nil
	}
}

func validateAccount(db *gorm.DB, account models.Account) error {
	if account.Name == "" {
		return invalid("name is required")
	}
	if !account.Kind.Valid() {
		return invalid("unknown account kind %q", account.Kind)
	}
	if account.ContributionTypeID != nil {
		if account.Kind.IsLiability() {
			return invalid("liability accounts cannot have a contribution type")
		}
		var contributionType models.ContributionType
		if err := db.First(&contributionType, *account.ContributionTypeID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return invalid("contribution type not found")
			}
			return err
		}
	}
	return nil
}

func accountExists(db *gorm.DB, id uint) error {
	var account models.Account
	return db.Select("id").First(&account, id).Error
}

// ---- Balances ----

func ListBalances(db *gorm.DB, accountID uint) ([]models.AccountBalance, error) {
	if err := accountExists(db, accountID); err != nil {
		return nil, err
	}
	var balances []models.AccountBalance
	if err := db.Where("account_id = ?", accountID).Order("date DESC").Find(&balances).Error; err != nil {
		return nil, err
	}
	return balances, nil
}

// SetBalance records the balance of an account on a day, replacing any
// existing snapshot for that day.
func SetBalance(db *gorm.DB, accountID uint, date time.Time, balance float64) error {
	if err := accountExists(db, accountID); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		return setBalance(tx, accountID, date, balance)
	})
}

type BalanceEntry struct {
	AccountID uint    `json:"account_id"`
	Balance   float64 `json:"balance"`
}

// SetBalances records several balances for the same day at once, e.g. a
// monthly check-in of every account.
func SetBalances(db *gorm.DB, date time.Time, entries []BalanceEntry) error {
	if len(entries) == 0 {
		return invalid("no balances given")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			if err := accountExists(tx, entry.AccountID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return invalid("account %d not found", entry.AccountID)
				}
				return err
			}
			if err := setBalance(tx, entry.AccountID, date, entry.Balance); err != nil {
				return err
			}
		}
		return nil
	})
}

func setBalance(tx *gorm.DB, accountID uint, date time.Time, balance float64) error {
	snapshot := models.AccountBalance{AccountID: accountID, Date: date, Balance: roundCents(balance)}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"balance", "updated_at"}),
	}).Create(&snapshot).Error; err != nil {
		return err
	}
	return refreshCachedBalance(tx, accountID)
}

func DeleteBalance(db *gorm.DB, accountID, balanceID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		result := tx.Unscoped().Where("account_id = ?", accountID).Delete(&models.AccountBalance{}, balanceID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return refreshCachedBalance(tx, accountID)
	})
}

func refreshCachedBalance(tx *gorm.DB, accountID uint) error {
	var latest models.AccountBalance
	balance := 0.0
	err := tx.Where("account_id = ?", accountID).Order("date DESC").First(&latest).Error
	if err == nil {
		balance = latest.Balance
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Model(&models.Account{}).Where("id = ?", accountID).Update("current_balance", balance).Error
}

// ---- Transactions ----

func ListTransactions(db *gorm.DB, accountID uint) ([]models.AccountTransaction, error) {
	if err := accountExists(db, accountID); err != nil {
		return nil, err
	}
	var transactions []models.AccountTransaction
	if err := db.Where("account_id = ?", accountID).Order("date DESC, id DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

type TransactionInput struct {
	Kind   models.TransactionKind `json:"kind"`
	Amount float64                `json:"amount"`
	Date   time.Time              `json:"-"`
	Note   string                 `json:"note"`
	// UpdateBalance also moves the account balance by the amount (deposit adds,
	// withdrawal subtracts), recorded as a snapshot on the transaction's date.
	// Both writes happen in one DB transaction, so a failure leaves neither.
	UpdateBalance bool `json:"update_balance"`
}

func CreateTransaction(db *gorm.DB, accountID uint, input TransactionInput) (*models.AccountTransaction, error) {
	if input.Kind == "" {
		input.Kind = models.TransactionDeposit
	}
	if !input.Kind.Valid() {
		return nil, invalid("kind must be deposit or withdrawal")
	}
	if input.Amount <= 0 {
		return nil, invalid("amount must be positive")
	}
	transaction := models.AccountTransaction{
		AccountID: accountID,
		Kind:      input.Kind,
		Amount:    roundCents(input.Amount),
		Date:      input.Date,
		Note:      strings.TrimSpace(input.Note),
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := accountExists(tx, accountID); err != nil {
			return err
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}
		if !input.UpdateBalance {
			return nil
		}
		return applyTransactionToBalance(tx, accountID, transaction)
	})
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// applyTransactionToBalance records latest balance +/- the transaction amount
// as a snapshot on the transaction's date. Only allowed on or after the latest
// snapshot: a back-dated adjustment would be overwritten by the later one.
func applyTransactionToBalance(tx *gorm.DB, accountID uint, transaction models.AccountTransaction) error {
	var latest models.AccountBalance
	current := 0.0
	err := tx.Where("account_id = ?", accountID).Order("date DESC").First(&latest).Error
	switch {
	case err == nil:
		if transaction.Date.Format(time.DateOnly) < latest.Date.Format(time.DateOnly) {
			return invalid("cannot update the balance from a transaction before the latest balance (%s)", latest.Date.Format(time.DateOnly))
		}
		current = latest.Balance
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return err
	}
	delta := transaction.Amount
	if transaction.Kind == models.TransactionWithdrawal {
		delta = -delta
	}
	return setBalance(tx, accountID, transaction.Date, current+delta)
}

func DeleteTransaction(db *gorm.DB, accountID, transactionID uint) error {
	result := db.Where("account_id = ?", accountID).Delete(&models.AccountTransaction{}, transactionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
