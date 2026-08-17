package money

import (
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
	return h.db.AutoMigrate(
		&models.InvestmentAccountType{},
		&models.InvestmentAccount{},
		&models.InvestmentDeposit{},
		&models.ContributionRule{},
	)
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	group := router.Group("/money", authMiddleware)
	controller.RegisterInvestmentRoutes(group.Group("/investments"), h.db)
}
