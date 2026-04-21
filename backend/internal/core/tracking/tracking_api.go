package tracking

import (
	"be-simpletracker/internal/core/tracking/models"
	"be-simpletracker/internal/core/tracking/routes"

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
	return h.db.AutoMigrate(&models.BodyWeightLog{}, &models.StepLog{})
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	group := router.Group("/tracking", authMiddleware)
	routes.RegisterTrackingRoutes(group, h.db)
}
