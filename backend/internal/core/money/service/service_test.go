package service

import (
	"be-simpletracker/internal/database"
	"context"
	"errors"
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
		&models.ContributionType{},
		&models.Account{},
		&models.AccountBalance{},
		&models.AccountTransaction{},
		&models.ContributionLimit{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.Local)
}

func mustAccount(t *testing.T, db *gorm.DB, input AccountInput, balance *float64, date time.Time) *models.Account {
	t.Helper()
	account, err := CreateAccount(database.WithDB(context.Background(), db), input, balance, date)
	if err != nil {
		t.Fatal(err)
	}
	return account
}

func TestTFSAPresetSharesRoomAcrossAccounts(t *testing.T) {
	db := setupTestDB(t)
	birthYear := 2004 // eligible from 2022
	tfsa, err := CreateContributionTypeFromPreset(database.WithDB(context.Background(), db), PresetInput{Key: "tfsa"})
	if err != nil {
		t.Fatal(err)
	}
	first := mustAccount(t, db, AccountInput{Name: "Wealthsimple TFSA", Kind: models.AccountKindInvestment, ContributionTypeID: &tfsa.ID}, nil, time.Now())
	second := mustAccount(t, db, AccountInput{Name: "Bank TFSA", Kind: models.AccountKindSavings, ContributionTypeID: &tfsa.ID}, nil, time.Now())

	for _, tx := range []struct {
		account uint
		kind    models.TransactionKind
		amount  float64
		date    time.Time
	}{
		{first.ID, models.TransactionDeposit, 10_000, day(2023, time.March, 1)},
		{second.ID, models.TransactionDeposit, 2_000, day(2025, time.June, 1)},
		{first.ID, models.TransactionWithdrawal, 3_000, day(2025, time.July, 1)},
	} {
		if _, err := CreateTransaction(database.WithDB(context.Background(), db), tx.account, TransactionInput{Kind: tx.kind, Amount: tx.amount, Date: tx.date}); err != nil {
			t.Fatal(err)
		}
	}

	summary, err := GetContributionType(database.WithDB(context.Background(), db), tfsa.ID, &birthYear, day(2026, time.November, 15))
	if err != nil {
		t.Fatal(err)
	}
	if summary.Room == nil {
		t.Fatalf("expected room, got issue %q", summary.RoomIssue)
	}
	// Earned: 2022 6000 + 2023 6500 + 2024-2026 7000*3 = 33500
	// Contributed 12000, 3000 withdrawn in 2025 restored in 2026.
	if summary.Room.StartYear != 2022 {
		t.Fatalf("start year = %d, want 2022", summary.Room.StartYear)
	}
	if summary.Room.Available != 33_500-12_000+3_000 {
		t.Fatalf("available = %v, want 24500", summary.Room.Available)
	}
	if len(summary.Accounts) != 2 {
		t.Fatalf("accounts = %d, want 2", len(summary.Accounts))
	}

	// Deleting an account removes its contributions from the room.
	if err := DeleteAccount(database.WithDB(context.Background(), db), second.ID); err != nil {
		t.Fatal(err)
	}
	summary, err = GetContributionType(database.WithDB(context.Background(), db), tfsa.ID, &birthYear, day(2026, time.November, 15))
	if err != nil {
		t.Fatal(err)
	}
	if summary.Room.Available != 33_500-10_000+3_000 {
		t.Fatalf("available after delete = %v, want 26500", summary.Room.Available)
	}
}

func TestRoomIssueWhenNotConfigured(t *testing.T) {
	db := setupTestDB(t)
	age := 18
	if _, err := CreateContributionType(database.WithDB(context.Background(), db), ContributionTypeInput{Name: "Custom", EligibilityAge: &age}); err != nil {
		t.Fatal(err)
	}
	types, err := ListContributionTypes(database.WithDB(context.Background(), db), nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if types[0].Room != nil || types[0].RoomIssue == "" {
		t.Fatalf("expected a room issue, got %+v", types[0])
	}
}

func TestContributionTypeValidation(t *testing.T) {
	db := setupTestDB(t)
	if _, err := CreateContributionTypeFromPreset(database.WithDB(context.Background(), db), PresetInput{Key: "tfsa"}); err != nil {
		t.Fatal(err)
	}
	var validation ValidationError
	if _, err := CreateContributionType(database.WithDB(context.Background(), db), ContributionTypeInput{Name: "tfsa"}); !errors.As(err, &validation) {
		t.Fatalf("expected duplicate name to be rejected, got %v", err)
	}
	year := 2025
	if _, err := CreateContributionType(database.WithDB(context.Background(), db), ContributionTypeInput{Name: "X", OpeningRoomYear: &year}); !errors.As(err, &validation) {
		t.Fatalf("expected opening year without amount to be rejected, got %v", err)
	}
	if _, err := CreateContributionTypeFromPreset(database.WithDB(context.Background(), db), PresetInput{Key: "nope"}); !errors.As(err, &validation) {
		t.Fatalf("expected unknown preset to be rejected, got %v", err)
	}
}

func TestContributionTypeInUseCannotBeDeleted(t *testing.T) {
	db := setupTestDB(t)
	rrsp, err := CreateContributionTypeFromPreset(database.WithDB(context.Background(), db), PresetInput{Key: "rrsp"})
	if err != nil {
		t.Fatal(err)
	}
	account := mustAccount(t, db, AccountInput{Name: "RRSP", Kind: models.AccountKindInvestment, ContributionTypeID: &rrsp.ID}, nil, time.Now())
	if err := DeleteContributionType(database.WithDB(context.Background(), db), rrsp.ID); err == nil {
		t.Fatal("expected delete to fail while assigned")
	}
	if _, err := UpdateAccount(database.WithDB(context.Background(), db), account.ID, AccountInput{Name: "RRSP", Kind: models.AccountKindInvestment}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteContributionType(database.WithDB(context.Background(), db), rrsp.ID); err != nil {
		t.Fatal(err)
	}
	// Name is free again after a delete.
	if _, err := CreateContributionTypeFromPreset(database.WithDB(context.Background(), db), PresetInput{Key: "rrsp"}); err != nil {
		t.Fatal(err)
	}
}

func TestLimitUpsertAndDelete(t *testing.T) {
	db := setupTestDB(t)
	fhsa, err := CreateContributionTypeFromPreset(database.WithDB(context.Background(), db), PresetInput{Key: "fhsa", StartYear: intPtr(2025)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := UpsertContributionLimit(database.WithDB(context.Background(), db), fhsa.ID, 2023, 9_000); err != nil {
		t.Fatal(err)
	}
	if err := DeleteContributionLimit(database.WithDB(context.Background(), db), fhsa.ID, 2023); err != nil {
		t.Fatal(err)
	}
	if _, err := UpsertContributionLimit(database.WithDB(context.Background(), db), fhsa.ID, 2023, 8_000); err != nil {
		t.Fatal(err)
	}
	summary, err := GetContributionType(database.WithDB(context.Background(), db), fhsa.ID, nil, day(2026, time.January, 1))
	if err != nil {
		t.Fatal(err)
	}
	if summary.Room.Available != 16_000 {
		t.Fatalf("available = %v, want 16000 (2025 + capped carry)", summary.Room.Available)
	}
}

func TestAccountBalancesAndSummary(t *testing.T) {
	db := setupTestDB(t)
	opening := 1_000.0
	account := mustAccount(t, db, AccountInput{Name: "Brokerage", Kind: models.AccountKindInvestment}, &opening, day(2026, time.January, 1))
	if err := SetBalance(database.WithDB(context.Background(), db), account.ID, day(2026, time.March, 1), 1_500); err != nil {
		t.Fatal(err)
	}
	// Same day replaces the snapshot.
	if err := SetBalance(database.WithDB(context.Background(), db), account.ID, day(2026, time.March, 1), 1_600); err != nil {
		t.Fatal(err)
	}
	// Back-dated snapshot does not become the current balance.
	if err := SetBalance(database.WithDB(context.Background(), db), account.ID, day(2025, time.December, 1), 900); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Amount: 500, Date: day(2026, time.February, 1)}); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Kind: models.TransactionWithdrawal, Amount: 100, Date: day(2026, time.February, 2)}); err != nil {
		t.Fatal(err)
	}

	summary, err := GetAccount(database.WithDB(context.Background(), db), account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.CurrentBalance != 1_600 || summary.PreviousBalance == nil || *summary.PreviousBalance != 1_000 {
		t.Fatalf("balance = %v previous = %v, want 1600 and 1000", summary.CurrentBalance, summary.PreviousBalance)
	}
	if summary.NetContributions != 400 || summary.Gain != 1_200 || summary.TransactionCount != 2 {
		t.Fatalf("summary = %+v, want net 400, gain 1200, 2 transactions", summary)
	}

	balances, err := ListBalances(database.WithDB(context.Background(), db), account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 3 {
		t.Fatalf("balances = %d, want 3", len(balances))
	}
	if err := DeleteBalance(database.WithDB(context.Background(), db), account.ID, balances[0].ID); err != nil {
		t.Fatal(err)
	}
	var cached models.Account
	db.First(&cached, account.ID)
	if cached.CurrentBalance != 1_000 {
		t.Fatalf("cached balance after delete = %v, want 1000", cached.CurrentBalance)
	}
}

func TestAccountValidation(t *testing.T) {
	db := setupTestDB(t)
	var validation ValidationError
	if _, err := CreateAccount(database.WithDB(context.Background(), db), AccountInput{Name: "", Kind: models.AccountKindChequing}, nil, time.Now()); !errors.As(err, &validation) {
		t.Fatalf("expected empty name to be rejected, got %v", err)
	}
	if _, err := CreateAccount(database.WithDB(context.Background(), db), AccountInput{Name: "X", Kind: "boat"}, nil, time.Now()); !errors.As(err, &validation) {
		t.Fatalf("expected unknown kind to be rejected, got %v", err)
	}
	missing := uint(999)
	if _, err := CreateAccount(database.WithDB(context.Background(), db), AccountInput{Name: "X", Kind: models.AccountKindInvestment, ContributionTypeID: &missing}, nil, time.Now()); !errors.As(err, &validation) {
		t.Fatalf("expected missing contribution type to be rejected, got %v", err)
	}
	tfsa, _ := CreateContributionTypeFromPreset(database.WithDB(context.Background(), db), PresetInput{Key: "tfsa"})
	if _, err := CreateAccount(database.WithDB(context.Background(), db), AccountInput{Name: "Card", Kind: models.AccountKindCreditCard, ContributionTypeID: &tfsa.ID}, nil, time.Now()); !errors.As(err, &validation) {
		t.Fatalf("expected liability with contribution type to be rejected, got %v", err)
	}
	account := mustAccount(t, db, AccountInput{Name: "Chequing", Kind: models.AccountKindChequing}, nil, time.Now())
	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Amount: 0, Date: time.Now()}); !errors.As(err, &validation) {
		t.Fatalf("expected zero amount to be rejected, got %v", err)
	}
	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Kind: "refund", Amount: 5, Date: time.Now()}); !errors.As(err, &validation) {
		t.Fatalf("expected unknown transaction kind to be rejected, got %v", err)
	}
}

func TestArchiveHidesAccountAndNetWorthHistory(t *testing.T) {
	db := setupTestDB(t)
	chequingBalance := 2_000.0
	cardBalance := 500.0
	chequing := mustAccount(t, db, AccountInput{Name: "Chequing", Kind: models.AccountKindChequing}, &chequingBalance, day(2026, time.January, 1))
	card := mustAccount(t, db, AccountInput{Name: "Visa", Kind: models.AccountKindCreditCard}, &cardBalance, day(2026, time.January, 1))
	if err := SetBalances(database.WithDB(context.Background(), db), day(2026, time.February, 1), []BalanceEntry{
		{AccountID: chequing.ID, Balance: 2_500},
		{AccountID: card.ID, Balance: 300},
	}); err != nil {
		t.Fatal(err)
	}
	if err := SetBalances(database.WithDB(context.Background(), db), day(2026, time.March, 1), []BalanceEntry{{AccountID: 999, Balance: 1}}); err == nil {
		t.Fatal("expected unknown account in bulk update to fail")
	}

	history, err := NetWorthHistory(database.WithDB(context.Background(), db))
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("history points = %d, want 2", len(history))
	}
	if history[0].NetWorth != 1_500 || history[1].NetWorth != 2_200 || history[1].Liabilities != 300 {
		t.Fatalf("history = %+v, want 1500 then 2200 with 300 liabilities", history)
	}

	if _, err := UpdateAccount(database.WithDB(context.Background(), db), card.ID, AccountInput{Name: "Visa", Kind: models.AccountKindCreditCard, Archived: true}); err != nil {
		t.Fatal(err)
	}
	active, err := ListAccounts(database.WithDB(context.Background(), db), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].ID != chequing.ID {
		t.Fatalf("active accounts = %+v, want only chequing", active)
	}
	all, err := ListAccounts(database.WithDB(context.Background(), db), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("all accounts = %d, want 2", len(all))
	}
	if !all[1].IsLiability {
		t.Fatal("expected credit card to be a liability")
	}

	// Past points still count the card; a later point after archiving does not.
	if err := SetBalance(database.WithDB(context.Background(), db), chequing.ID, time.Now().AddDate(0, 0, 1), 3_000); err != nil {
		t.Fatal(err)
	}
	history, err = NetWorthHistory(database.WithDB(context.Background(), db))
	if err != nil {
		t.Fatal(err)
	}
	if history[1].Liabilities != 300 {
		t.Fatalf("past point liabilities = %v, want 300", history[1].Liabilities)
	}
	last := history[len(history)-1]
	if last.Liabilities != 0 || last.NetWorth != 3_000 {
		t.Fatalf("last point = %+v, want card excluded after archiving", last)
	}
}

func TestTransactionUpdatesBalanceAtomically(t *testing.T) {
	db := setupTestDB(t)
	opening := 1_000.0
	account := mustAccount(t, db, AccountInput{Name: "Savings", Kind: models.AccountKindCash}, &opening, day(2026, time.March, 1))

	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Amount: 250.5, Date: day(2026, time.March, 5), UpdateBalance: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Kind: models.TransactionWithdrawal, Amount: 50.5, Date: day(2026, time.March, 5), UpdateBalance: true}); err != nil {
		t.Fatal(err)
	}
	var cached models.Account
	db.First(&cached, account.ID)
	if cached.CurrentBalance != 1_200 {
		t.Fatalf("balance = %v, want 1200", cached.CurrentBalance)
	}

	// Before the latest snapshot: rejected, and the transaction is not saved either.
	var validation ValidationError
	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Amount: 10, Date: day(2026, time.March, 2), UpdateBalance: true}); !errors.As(err, &validation) {
		t.Fatalf("back-dated update err = %v, want validation error", err)
	}
	var count int64
	db.Model(&models.AccountTransaction{}).Where("account_id = ?", account.ID).Count(&count)
	if count != 2 {
		t.Fatalf("transactions = %d, want 2 (rejected one rolled back)", count)
	}

	// Without UpdateBalance the balance is untouched.
	if _, err := CreateTransaction(database.WithDB(context.Background(), db), account.ID, TransactionInput{Amount: 10, Date: day(2026, time.March, 2)}); err != nil {
		t.Fatal(err)
	}
	db.First(&cached, account.ID)
	if cached.CurrentBalance != 1_200 {
		t.Fatalf("balance = %v, want 1200", cached.CurrentBalance)
	}
}
