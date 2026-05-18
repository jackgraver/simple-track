package service

import (
	finance_model "be-simpletracker/internal/core/finance/model"
	finance_repository "be-simpletracker/internal/core/finance/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

	
func ListAccounts() ([]finance_model.Account, error) {
	return finance_repository.ListAccounts()
}

func CreateAccount(name string, balance float32) (*finance_model.Account, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	a := &finance_model.Account{Name: name, Balance: balance}
	if err := finance_repository.CreateAccount(a); err != nil {
		return nil, err
	}
	return a, nil
}

func ListTransactions() ([]finance_model.Transaction, error) {
	return finance_repository.ListTransactions()
}

func ListCategories() ([]finance_model.TransactionCategory, error) {
	return finance_repository.ListCategories()
}

func CreateTransaction(accountID uint, amount float32, at time.Time, categoryID uint) (*finance_model.Transaction, error) {
	if accountID == 0 {
		return nil, errors.New("account_id is required")
	}
	if categoryID == 0 {
		return nil, errors.New("category_id is required")
	}
	if _, err := finance_repository.GetCategory(categoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	if _, err := finance_repository.GetAccount(accountID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	t := &finance_model.Transaction{
		AccountID:  accountID,
		Amount:     amount,
		Date:       at,
		CategoryID: categoryID,
	}
	if err := finance_repository.CreateTransaction(t); err != nil {
		return nil, err
	}
	return t, nil
}
