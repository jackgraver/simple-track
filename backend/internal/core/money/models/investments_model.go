package models

import (
	"time"

	"gorm.io/gorm"
)

type InvestmentAccount struct {
	gorm.Model
	Name                    string                 `json:"name" gorm:"not null"`
	InvestmentAccountTypeID *uint                  `json:"investment_account_type_id"`
	InvestmentAccountType   *InvestmentAccountType `json:"investment_account_type,omitempty" gorm:"foreignKey:InvestmentAccountTypeID"`
	CurrentBalance          float64                `json:"current_balance" gorm:"type:numeric(14,2);not null;default:0"`
	Deposits                []InvestmentDeposit    `json:"-" gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
}

func (InvestmentAccount) TableName() string { return "investment_accounts" }

type InvestmentAccountType struct {
	gorm.Model
	Name                 string             `json:"name" gorm:"not null;uniqueIndex"`
	ContributionStartAge *int               `json:"contribution_start_age,omitempty"`
	Rules                []ContributionRule `json:"rules,omitempty" gorm:"foreignKey:InvestmentAccountTypeID;constraint:OnDelete:CASCADE"`
}

func (InvestmentAccountType) TableName() string { return "investment_account_types" }

type InvestmentDeposit struct {
	gorm.Model
	AccountID uint      `json:"account_id" gorm:"not null;index"`
	Amount    float64   `json:"amount" gorm:"type:numeric(14,2);not null"`
	Date      time.Time `json:"date" gorm:"not null;index"`
}

func (InvestmentDeposit) TableName() string { return "investment_deposits" }

type ContributionRule struct {
	gorm.Model
	InvestmentAccountTypeID uint    `json:"investment_account_type_id" gorm:"not null;uniqueIndex:idx_investment_account_type_contribution_rule,priority:1"`
	Year                    int     `json:"year" gorm:"not null;uniqueIndex:idx_investment_account_type_contribution_rule,priority:2"`
	AnnualLimit             float64 `json:"annual_limit" gorm:"type:numeric(14,2);not null"`
}

func (ContributionRule) TableName() string { return "investment_contribution_rules" }
