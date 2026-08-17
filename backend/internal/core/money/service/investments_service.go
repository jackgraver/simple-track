package service

import (
	"errors"
	"strings"
	"time"

	"be-simpletracker/internal/core/money/models"

	"gorm.io/gorm"
)

type ContributionStatus struct {
	Year        int     `json:"year"`
	AnnualLimit float64 `json:"annual_limit"`
	Contributed float64 `json:"contributed"`
	Remaining   float64 `json:"remaining"`
}

type ContributionRoom struct {
	EligibleFromYear int     `json:"eligible_from_year"`
	EarnedRoom       float64 `json:"earned_room"`
	Contributed      float64 `json:"contributed"`
	Remaining        float64 `json:"remaining"`
}

type InvestmentAccountSummary struct {
	models.InvestmentAccount
	TotalDeposits      float64              `json:"total_deposits"`
	Profit             float64              `json:"profit"`
	ContributionStatus []ContributionStatus `json:"contribution_status"`
	ContributionRoom   *ContributionRoom    `json:"contribution_room,omitempty"`
}

type InvestmentAccountTypeSummary struct {
	models.InvestmentAccountType
	ContributionStatus []ContributionStatus `json:"contribution_status"`
	ContributionRoom   *ContributionRoom    `json:"contribution_room,omitempty"`
}

func ListInvestmentAccountTypes(db *gorm.DB, birthYear *int) ([]InvestmentAccountTypeSummary, error) {
	var accountTypes []models.InvestmentAccountType
	if err := db.Preload("Rules", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("year DESC")
	}).Order("name ASC").Find(&accountTypes).Error; err != nil {
		return nil, err
	}
	summaries := make([]InvestmentAccountTypeSummary, 0, len(accountTypes))
	for _, accountType := range accountTypes {
		status, err := contributionStatus(db, accountType.ID, accountType.Rules)
		if err != nil {
			return nil, err
		}
		room, err := contributionRoom(db, accountType.ID, accountType.ContributionStartAge, birthYear, accountType.Rules)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, InvestmentAccountTypeSummary{
			InvestmentAccountType: accountType,
			ContributionStatus:    status,
			ContributionRoom:      room,
		})
	}
	return summaries, nil
}

func ListInvestmentAccounts(db *gorm.DB, birthYear *int) ([]InvestmentAccountSummary, error) {
	var accounts []models.InvestmentAccount
	if err := db.Preload("InvestmentAccountType").Order("name ASC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	summaries := make([]InvestmentAccountSummary, 0, len(accounts))
	for _, account := range accounts {
		summary, err := GetInvestmentAccountSummary(db, account.ID, birthYear)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, *summary)
	}
	return summaries, nil
}

func GetInvestmentAccountSummary(db *gorm.DB, id uint, birthYear *int) (*InvestmentAccountSummary, error) {
	var account models.InvestmentAccount
	if err := db.Preload("InvestmentAccountType.Rules", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("year DESC")
	}).First(&account, id).Error; err != nil {
		return nil, err
	}
	var totalDeposits float64
	if err := db.Model(&models.InvestmentDeposit{}).
		Where("account_id = ?", id).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalDeposits).Error; err != nil {
		return nil, err
	}
	status := []ContributionStatus{}
	var room *ContributionRoom
	if account.InvestmentAccountType != nil {
		var err error
		status, err = contributionStatus(db, account.InvestmentAccountType.ID, account.InvestmentAccountType.Rules)
		if err != nil {
			return nil, err
		}
		room, err = contributionRoom(db, account.InvestmentAccountType.ID, account.InvestmentAccountType.ContributionStartAge, birthYear, account.InvestmentAccountType.Rules)
		if err != nil {
			return nil, err
		}
	}
	return &InvestmentAccountSummary{
		InvestmentAccount:  account,
		TotalDeposits:      totalDeposits,
		Profit:             account.CurrentBalance - totalDeposits,
		ContributionStatus: status,
		ContributionRoom:   room,
	}, nil
}

func CreateInvestmentAccount(db *gorm.DB, name string, accountTypeID *uint, currentBalance float64) (*models.InvestmentAccount, error) {
	account := models.InvestmentAccount{
		Name:                    strings.TrimSpace(name),
		InvestmentAccountTypeID: accountTypeID,
		CurrentBalance:          currentBalance,
	}
	if err := validateAccount(db, account); err != nil {
		return nil, err
	}
	if err := db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func UpdateInvestmentAccount(db *gorm.DB, id uint, name *string, accountTypeID *uint, currentBalance *float64) (*models.InvestmentAccount, error) {
	var account models.InvestmentAccount
	if err := db.First(&account, id).Error; err != nil {
		return nil, err
	}
	if name != nil {
		account.Name = strings.TrimSpace(*name)
	}
	if accountTypeID != nil {
		account.InvestmentAccountTypeID = accountTypeID
	}
	if currentBalance != nil {
		account.CurrentBalance = *currentBalance
	}
	if err := validateAccount(db, account); err != nil {
		return nil, err
	}
	if err := db.Save(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func DeleteInvestmentAccount(db *gorm.DB, id uint) error {
	result := db.Delete(&models.InvestmentAccount{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ListInvestmentDeposits(db *gorm.DB, accountID uint) ([]models.InvestmentDeposit, error) {
	if err := accountExists(db, accountID); err != nil {
		return nil, err
	}
	var deposits []models.InvestmentDeposit
	if err := db.Where("account_id = ?", accountID).Order("date DESC, id DESC").Find(&deposits).Error; err != nil {
		return nil, err
	}
	return deposits, nil
}

func CreateInvestmentDeposit(db *gorm.DB, accountID uint, amount float64, date time.Time) (*models.InvestmentDeposit, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if err := accountExists(db, accountID); err != nil {
		return nil, err
	}
	deposit := models.InvestmentDeposit{AccountID: accountID, Amount: amount, Date: date}
	if err := db.Create(&deposit).Error; err != nil {
		return nil, err
	}
	return &deposit, nil
}

func DeleteInvestmentDeposit(db *gorm.DB, accountID, depositID uint) error {
	result := db.Where("account_id = ?", accountID).Delete(&models.InvestmentDeposit{}, depositID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func CreateInvestmentAccountType(db *gorm.DB, name string, contributionStartAge *int) (*models.InvestmentAccountType, error) {
	accountType := models.InvestmentAccountType{Name: strings.TrimSpace(name), ContributionStartAge: contributionStartAge}
	if err := validateAccountType(accountType); err != nil {
		return nil, err
	}
	if err := db.Create(&accountType).Error; err != nil {
		return nil, err
	}
	return &accountType, nil
}

func UpdateInvestmentAccountType(db *gorm.DB, id uint, name string, contributionStartAge *int) (*models.InvestmentAccountType, error) {
	var accountType models.InvestmentAccountType
	if err := db.First(&accountType, id).Error; err != nil {
		return nil, err
	}
	accountType.Name = strings.TrimSpace(name)
	if contributionStartAge != nil {
		accountType.ContributionStartAge = contributionStartAge
	}
	if err := validateAccountType(accountType); err != nil {
		return nil, err
	}
	if err := db.Save(&accountType).Error; err != nil {
		return nil, err
	}
	return &accountType, nil
}

func DeleteInvestmentAccountType(db *gorm.DB, id uint) error {
	var count int64
	if err := db.Model(&models.InvestmentAccount{}).Where("investment_account_type_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("account type is assigned to one or more accounts")
	}
	result := db.Delete(&models.InvestmentAccountType{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func UpsertContributionRule(db *gorm.DB, accountTypeID uint, year int, annualLimit float64) (*models.ContributionRule, error) {
	if year < 1900 || year > 9999 {
		return nil, errors.New("year must be between 1900 and 9999")
	}
	if annualLimit < 0 {
		return nil, errors.New("annual_limit must not be negative")
	}
	if err := accountTypeExists(db, accountTypeID); err != nil {
		return nil, err
	}
	var rule models.ContributionRule
	err := db.Where("investment_account_type_id = ? AND year = ?", accountTypeID, year).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rule = models.ContributionRule{InvestmentAccountTypeID: accountTypeID, Year: year, AnnualLimit: annualLimit}
		if err := db.Create(&rule).Error; err != nil {
			return nil, err
		}
		return &rule, nil
	}
	if err != nil {
		return nil, err
	}
	rule.AnnualLimit = annualLimit
	if err := db.Save(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func DeleteContributionRule(db *gorm.DB, accountTypeID uint, year int) error {
	result := db.Where("investment_account_type_id = ? AND year = ?", accountTypeID, year).Delete(&models.ContributionRule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func contributionStatus(db *gorm.DB, accountTypeID uint, rules []models.ContributionRule) ([]ContributionStatus, error) {
	status := make([]ContributionStatus, 0, len(rules))
	for _, rule := range rules {
		contributed, err := contributionTotal(db, accountTypeID, rule.Year)
		if err != nil {
			return nil, err
		}
		status = append(status, ContributionStatus{
			Year:        rule.Year,
			AnnualLimit: rule.AnnualLimit,
			Contributed: contributed,
			Remaining:   rule.AnnualLimit - contributed,
		})
	}
	return status, nil
}

func contributionRoom(db *gorm.DB, accountTypeID uint, contributionStartAge, birthYear *int, rules []models.ContributionRule) (*ContributionRoom, error) {
	if contributionStartAge == nil || birthYear == nil {
		return nil, nil
	}
	eligibleFromYear := *birthYear + *contributionStartAge
	currentYear := time.Now().Year()
	var earnedRoom float64
	for _, rule := range rules {
		if rule.Year >= eligibleFromYear && rule.Year <= currentYear {
			earnedRoom += rule.AnnualLimit
		}
	}
	contributed, err := contributionTotalThroughYear(db, accountTypeID, eligibleFromYear, currentYear)
	if err != nil {
		return nil, err
	}
	return &ContributionRoom{
		EligibleFromYear: eligibleFromYear,
		EarnedRoom:       earnedRoom,
		Contributed:      contributed,
		Remaining:        earnedRoom - contributed,
	}, nil
}

func contributionTotal(db *gorm.DB, accountTypeID uint, year int) (float64, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	var total float64
	err := db.Model(&models.InvestmentDeposit{}).
		Joins("JOIN investment_accounts ON investment_accounts.id = investment_deposits.account_id").
		Where("investment_accounts.investment_account_type_id = ? AND investment_deposits.date >= ? AND investment_deposits.date < ?", accountTypeID, start, end).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func contributionTotalThroughYear(db *gorm.DB, accountTypeID uint, startYear, endYear int) (float64, error) {
	start := time.Date(startYear, time.January, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(endYear+1, time.January, 1, 0, 0, 0, 0, time.Local)
	var total float64
	err := db.Model(&models.InvestmentDeposit{}).
		Joins("JOIN investment_accounts ON investment_accounts.id = investment_deposits.account_id").
		Where("investment_accounts.investment_account_type_id = ? AND investment_deposits.date >= ? AND investment_deposits.date < ?", accountTypeID, start, end).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func accountExists(db *gorm.DB, id uint) error {
	var account models.InvestmentAccount
	return db.First(&account, id).Error
}

func accountTypeExists(db *gorm.DB, id uint) error {
	var accountType models.InvestmentAccountType
	return db.First(&accountType, id).Error
}

func validateAccount(db *gorm.DB, account models.InvestmentAccount) error {
	if account.Name == "" {
		return errors.New("name is required")
	}
	if account.CurrentBalance < 0 {
		return errors.New("current_balance must not be negative")
	}
	if account.InvestmentAccountTypeID != nil {
		return accountTypeExists(db, *account.InvestmentAccountTypeID)
	}
	return nil
}

func validateAccountType(accountType models.InvestmentAccountType) error {
	if accountType.Name == "" {
		return errors.New("name is required")
	}
	if accountType.ContributionStartAge != nil && (*accountType.ContributionStartAge < 0 || *accountType.ContributionStartAge > 120) {
		return errors.New("contribution_start_age must be between 0 and 120")
	}
	return nil
}
