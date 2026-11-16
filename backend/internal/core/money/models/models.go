package models

import (
	"time"

	"gorm.io/gorm"
)

// AccountKind describes what an account is. Liability kinds (credit cards,
// loans) hold an amount owed and subtract from net worth.
type AccountKind string

const (
	AccountKindChequing   AccountKind = "chequing"
	AccountKindSavings    AccountKind = "savings"
	AccountKindCash       AccountKind = "cash"
	AccountKindInvestment AccountKind = "investment"
	AccountKindProperty   AccountKind = "property"
	AccountKindCreditCard AccountKind = "credit_card"
	AccountKindLoan       AccountKind = "loan"
	AccountKindOther      AccountKind = "other"
)

var AccountKinds = []AccountKind{
	AccountKindChequing,
	AccountKindSavings,
	AccountKindCash,
	AccountKindInvestment,
	AccountKindProperty,
	AccountKindCreditCard,
	AccountKindLoan,
	AccountKindOther,
}

func (k AccountKind) Valid() bool {
	for _, kind := range AccountKinds {
		if kind == k {
			return true
		}
	}
	return false
}

func (k AccountKind) IsLiability() bool {
	return k == AccountKindCreditCard || k == AccountKindLoan
}

// Account is any account whose balance is tracked. The table name is kept
// from when the module only tracked investment accounts so existing data
// carries over.
type Account struct {
	gorm.Model
	Name               string            `json:"name" gorm:"not null"`
	Kind               AccountKind       `json:"kind" gorm:"type:varchar(32);not null;default:'investment'"`
	Institution        string            `json:"institution" gorm:"not null;default:''"`
	Notes              string            `json:"notes" gorm:"not null;default:''"`
	ContributionTypeID *uint             `json:"contribution_type_id" gorm:"column:investment_account_type_id"`
	ContributionType   *ContributionType `json:"contribution_type,omitempty" gorm:"foreignKey:ContributionTypeID"`
	// CurrentBalance caches the most recent AccountBalance snapshot.
	CurrentBalance float64              `json:"current_balance" gorm:"type:numeric(14,2);not null;default:0"`
	ArchivedAt     *time.Time           `json:"archived_at"`
	Balances       []AccountBalance     `json:"-" gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
	Transactions   []AccountTransaction `json:"-" gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
}

func (Account) TableName() string { return "investment_accounts" }

// AccountBalance is the balance of an account on a given day. For liability
// accounts it is the amount owed.
type AccountBalance struct {
	gorm.Model
	AccountID uint      `json:"account_id" gorm:"not null;uniqueIndex:idx_account_balance_day,priority:1"`
	Date      time.Time `json:"date" gorm:"not null;uniqueIndex:idx_account_balance_day,priority:2"`
	Balance   float64   `json:"balance" gorm:"type:numeric(14,2);not null"`
}

func (AccountBalance) TableName() string { return "account_balances" }

type TransactionKind string

const (
	TransactionDeposit    TransactionKind = "deposit"
	TransactionWithdrawal TransactionKind = "withdrawal"
)

func (k TransactionKind) Valid() bool {
	return k == TransactionDeposit || k == TransactionWithdrawal
}

// AccountTransaction is money moved into (deposit) or out of (withdrawal) an
// account. Deposits into an account with a contribution type use up room.
type AccountTransaction struct {
	gorm.Model
	AccountID uint            `json:"account_id" gorm:"not null;index"`
	Kind      TransactionKind `json:"kind" gorm:"type:varchar(16);not null;default:'deposit'"`
	Amount    float64         `json:"amount" gorm:"type:numeric(14,2);not null"`
	Date      time.Time       `json:"date" gorm:"not null;index"`
	Note      string          `json:"note" gorm:"not null;default:''"`
}

func (AccountTransaction) TableName() string { return "investment_deposits" }

// ContributionType is a registered account type such as a TFSA, RRSP or FHSA
// whose contribution room is shared by every account assigned to it.
type ContributionType struct {
	gorm.Model
	Name string `json:"name" gorm:"not null;uniqueIndex"`
	// StartYear is the first year room accrues. When nil it is derived from
	// the user's birth year plus EligibilityAge, or the first limit year.
	StartYear      *int `json:"start_year" gorm:"column:contribution_start_year"`
	EligibilityAge *int `json:"eligibility_age"`
	// CarryForwardCap limits how much unused room rolls into the next year.
	// nil means unlimited, 0 means unused room is lost.
	CarryForwardCap *float64 `json:"carry_forward_cap" gorm:"type:numeric(14,2)"`
	// LifetimeLimit caps total contributions over all years.
	LifetimeLimit *float64 `json:"lifetime_limit" gorm:"type:numeric(14,2)"`
	// WithdrawalsRestoreRoom adds withdrawals back as room on January 1 of
	// the following year (TFSA behaviour).
	WithdrawalsRestoreRoom bool `json:"withdrawals_restore_room" gorm:"not null;default:false"`
	// OpeningRoomYear/OpeningRoom anchor the calculation to a known figure,
	// e.g. "available room on Jan 1, 2025" from CRA. Earlier history is then
	// ignored.
	OpeningRoomYear *int                `json:"opening_room_year"`
	OpeningRoom     *float64            `json:"opening_room" gorm:"type:numeric(14,2)"`
	Limits          []ContributionLimit `json:"limits" gorm:"foreignKey:ContributionTypeID;constraint:OnDelete:CASCADE"`
}

func (ContributionType) TableName() string { return "investment_account_types" }

// ContributionLimit is the new room granted in Year. It keeps applying to
// later years until another limit replaces it.
type ContributionLimit struct {
	gorm.Model
	ContributionTypeID uint    `json:"contribution_type_id" gorm:"column:investment_account_type_id;not null;uniqueIndex:idx_investment_account_type_contribution_rule,priority:1"`
	Year               int     `json:"year" gorm:"not null;uniqueIndex:idx_investment_account_type_contribution_rule,priority:2"`
	AnnualLimit        float64 `json:"annual_limit" gorm:"type:numeric(14,2);not null"`
}

func (ContributionLimit) TableName() string { return "investment_contribution_rules" }
