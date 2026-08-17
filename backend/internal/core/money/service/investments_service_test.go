package service

import (
	"testing"
	"time"

	"be-simpletracker/internal/core/money/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.InvestmentAccountType{},
		&models.InvestmentAccount{},
		&models.InvestmentDeposit{},
		&models.ContributionRule{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestInvestmentAccountSummaryTracksProfitAndContributionRoom(t *testing.T) {
	db := setupTestDB(t)
	startAge := 18
	birthYear := 2000
	accountType, err := CreateInvestmentAccountType(db, "TFSA", &startAge)
	if err != nil {
		t.Fatal(err)
	}
	account, err := CreateInvestmentAccount(db, "Wealthsimple TFSA", &accountType.ID, 1_200)
	if err != nil {
		t.Fatal(err)
	}
	secondAccount, err := CreateInvestmentAccount(db, "Bank TFSA", &accountType.ID, 400)
	if err != nil {
		t.Fatal(err)
	}
	for _, deposit := range []struct {
		amount float64
		date   time.Time
	}{
		{amount: 500, date: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.Local)},
		{amount: 300, date: time.Date(2026, time.March, 10, 0, 0, 0, 0, time.Local)},
		{amount: 100, date: time.Date(2025, time.December, 31, 0, 0, 0, 0, time.Local)},
	} {
		if _, err := CreateInvestmentDeposit(db, account.ID, deposit.amount, deposit.date); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := CreateInvestmentDeposit(db, secondAccount.ID, 200, time.Date(2026, time.April, 8, 0, 0, 0, 0, time.Local)); err != nil {
		t.Fatal(err)
	}
	if _, err := UpsertContributionRule(db, accountType.ID, 2026, 7_000); err != nil {
		t.Fatal(err)
	}
	if _, err := UpsertContributionRule(db, accountType.ID, 2025, 6_000); err != nil {
		t.Fatal(err)
	}
	if _, err := UpsertContributionRule(db, accountType.ID, 2017, 5_000); err != nil {
		t.Fatal(err)
	}
	summary, err := GetInvestmentAccountSummary(db, account.ID, &birthYear)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalDeposits != 900 {
		t.Fatalf("total deposits = %v, want 900", summary.TotalDeposits)
	}
	if summary.Profit != 300 {
		t.Fatalf("profit = %v, want 300", summary.Profit)
	}
	if len(summary.ContributionStatus) != 3 {
		t.Fatalf("contribution statuses = %d, want 3", len(summary.ContributionStatus))
	}
	status := summary.ContributionStatus[0]
	if status.Contributed != 1_000 || status.Remaining != 6_000 {
		t.Fatalf("contribution status = %#v, want contributed 1000 and remaining 6000", status)
	}
	if summary.ContributionRoom == nil {
		t.Fatal("expected cumulative contribution room")
	}
	if summary.ContributionRoom.EligibleFromYear != 2018 || summary.ContributionRoom.EarnedRoom != 13_000 || summary.ContributionRoom.Contributed != 1_100 || summary.ContributionRoom.Remaining != 11_900 {
		t.Fatalf("contribution room = %#v, want eligibility 2018, earned 13000, contributed 1100, remaining 11900", summary.ContributionRoom)
	}
}

func TestInvestmentAccountValidation(t *testing.T) {
	db := setupTestDB(t)
	if _, err := CreateInvestmentAccount(db, "", nil, 0); err == nil {
		t.Fatal("expected empty name to be rejected")
	}
	invalidTypeID := uint(999)
	if _, err := CreateInvestmentAccount(db, "Invalid", &invalidTypeID, 0); err == nil {
		t.Fatal("expected missing account type to be rejected")
	}
	if _, err := CreateInvestmentAccount(db, "Negative", nil, -1); err == nil {
		t.Fatal("expected negative balance to be rejected")
	}
}
