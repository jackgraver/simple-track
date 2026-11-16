package money

import (
	"testing"
	"time"

	"be-simpletracker/internal/core/money/models"
	"be-simpletracker/internal/core/money/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Snapshot of the schema before accounts were generalised, used to check that
// Migrate upgrades existing data in place.
type legacyAccountType struct {
	gorm.Model
	Name                  string `gorm:"not null;uniqueIndex"`
	ContributionStartYear *int
}

func (legacyAccountType) TableName() string { return "investment_account_types" }

type legacyAccount struct {
	gorm.Model
	Name                    string `gorm:"not null"`
	InvestmentAccountTypeID *uint
	CurrentBalance          float64 `gorm:"type:numeric(14,2);not null;default:0"`
}

func (legacyAccount) TableName() string { return "investment_accounts" }

type legacyDeposit struct {
	gorm.Model
	AccountID uint      `gorm:"not null;index"`
	Amount    float64   `gorm:"type:numeric(14,2);not null"`
	Date      time.Time `gorm:"not null;index"`
}

func (legacyDeposit) TableName() string { return "investment_deposits" }

type legacyRule struct {
	gorm.Model
	InvestmentAccountTypeID uint    `gorm:"not null;uniqueIndex:idx_investment_account_type_contribution_rule,priority:1"`
	Year                    int     `gorm:"not null;uniqueIndex:idx_investment_account_type_contribution_rule,priority:2"`
	AnnualLimit             float64 `gorm:"type:numeric(14,2);not null"`
}

func (legacyRule) TableName() string { return "investment_contribution_rules" }

func TestMigrateUpgradesLegacyData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&legacyAccountType{}, &legacyAccount{}, &legacyDeposit{}, &legacyRule{}); err != nil {
		t.Fatal(err)
	}
	startYear := 2025
	accountType := legacyAccountType{Name: "TFSA", ContributionStartYear: &startYear}
	db.Create(&accountType)
	db.Create(&legacyRule{InvestmentAccountTypeID: accountType.ID, Year: 2025, AnnualLimit: 7000})
	account := legacyAccount{Name: "Wealthsimple", InvestmentAccountTypeID: &accountType.ID, CurrentBalance: 1234.56}
	db.Create(&account)
	db.Create(&legacyDeposit{AccountID: account.ID, Amount: 1000, Date: time.Date(2025, 5, 1, 0, 0, 0, 0, time.Local)})
	// A soft-deleted type would otherwise block re-creating the same name.
	deleted := legacyAccountType{Name: "FHSA"}
	db.Create(&deleted)
	db.Delete(&deleted)

	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	// Running twice must be safe.
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	summary, err := service.GetAccount(db, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Kind != models.AccountKindInvestment {
		t.Fatalf("kind = %q, want investment", summary.Kind)
	}
	if summary.CurrentBalance != 1234.56 || summary.BalanceDate == nil {
		t.Fatalf("balance = %v date = %v, want backfilled snapshot", summary.CurrentBalance, summary.BalanceDate)
	}
	if summary.TotalDeposits != 1000 || summary.ContributionType == nil || summary.ContributionType.Name != "TFSA" {
		t.Fatalf("summary = %+v, want deposits and TFSA type preserved", summary)
	}
	balances, _ := service.ListBalances(db, account.ID)
	if len(balances) != 1 {
		t.Fatalf("balances = %d, want exactly 1 after two migrations", len(balances))
	}

	types, err := service.ListContributionTypes(db, nil, time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 1 || types[0].Room == nil {
		t.Fatalf("types = %+v, want TFSA with room", types)
	}
	if types[0].Room.Available != 13000 || len(types[0].Limits) != 1 {
		t.Fatalf("room = %+v, want 13000 available from preserved rule", types[0].Room)
	}
	if _, err := service.CreateContributionTypeFromPreset(db, service.PresetInput{Key: "fhsa", StartYear: &startYear}); err != nil {
		t.Fatalf("expected FHSA name to be free after purge: %v", err)
	}
}
