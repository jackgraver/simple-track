package money

import (
	"time"

	"be-simpletracker/internal/core/money/controller"
	"be-simpletracker/internal/core/money/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Migrate() error {
	return Migrate(h.db)
}

// Migrate creates/updates the money tables. Table names are kept from the
// original investments-only schema, so existing accounts, deposits and
// contribution rules carry over.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.ContributionType{},
		&models.Account{},
		&models.AccountBalance{},
		&models.AccountTransaction{},
		&models.ContributionLimit{},
	); err != nil {
		return err
	}
	// Contribution types used to be soft-deleted, which blocks re-creating
	// one with the same name because of the unique index.
	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.ContributionLimit{}).Error; err != nil {
		return err
	}
	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.ContributionType{}).Error; err != nil {
		return err
	}
	return backfillBalanceHistory(db)
}

// backfillBalanceHistory gives every account without balance history a
// snapshot of its cached balance, dated when the account was last updated.
func backfillBalanceHistory(db *gorm.DB) error {
	var accounts []models.Account
	if err := db.Where("NOT EXISTS (SELECT 1 FROM account_balances WHERE account_balances.account_id = investment_accounts.id)").
		Find(&accounts).Error; err != nil {
		return err
	}
	for _, account := range accounts {
		updated := account.UpdatedAt.Local()
		snapshot := models.AccountBalance{
			AccountID: account.ID,
			Date:      time.Date(updated.Year(), updated.Month(), updated.Day(), 0, 0, 0, 0, time.Local),
			Balance:   account.CurrentBalance,
		}
		if err := db.Create(&snapshot).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	controller.RegisterRoutes(router.Group("/money", authMiddleware), h.db)
}
