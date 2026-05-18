package model

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Name    string  `json:"name"    gorm:"not null"`
	Balance float32 `json:"balance" gorm:"not null"`
}

type Transaction struct {
	gorm.Model
	AccountID   uint                `json:"account_id"   gorm:"not null"`
	Account     Account             `json:"account"`
	Amount      float32             `json:"amount"       gorm:"not null"`
	Date        time.Time           `json:"date"         gorm:"not null"`
	CategoryID  uint                `json:"category_id"  gorm:"not null"`
	Category    TransactionCategory `json:"category"`
}

type TransactionCategory struct {
	gorm.Model
	Name string `json:"name" gorm:"not null"`
}