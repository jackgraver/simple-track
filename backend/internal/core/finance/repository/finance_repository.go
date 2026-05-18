package repository

import (
	finance_model "be-simpletracker/internal/core/finance/model"
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
)

func conn() *gorm.DB {
	return database.GetDB()
}

func ListAccounts() ([]finance_model.Account, error) {
	var rows []finance_model.Account
	err := conn().Order("id asc").Find(&rows).Error
	return rows, err
}

func CreateAccount(a *finance_model.Account) error {
	return conn().Create(a).Error
}

func GetAccount(id uint) (*finance_model.Account, error) {
	var a finance_model.Account
	err := conn().First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func ListTransactions() ([]finance_model.Transaction, error) {
	var rows []finance_model.Transaction
	err := conn().Preload("Account").Preload("Category").Order("date desc, id desc").Find(&rows).Error
	return rows, err
}

func CreateTransaction(t *finance_model.Transaction) error {
	return conn().Create(t).Error
}

func UncategorizedCategoryID() (uint, error) {
	var cat finance_model.TransactionCategory
	err := conn().FirstOrCreate(&cat, finance_model.TransactionCategory{Name: "Uncategorized"}).Error
	return cat.ID, err
}

func ListCategories() ([]finance_model.TransactionCategory, error) {
	if _, err := UncategorizedCategoryID(); err != nil {
		return nil, err
	}
	var rows []finance_model.TransactionCategory
	err := conn().Order("name asc").Find(&rows).Error
	return rows, err
}

func GetCategory(id uint) (*finance_model.TransactionCategory, error) {
	var cat finance_model.TransactionCategory
	err := conn().First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}
