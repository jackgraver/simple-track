package tracking

import (
	"be-simpletracker/internal/core/tracking/missed"
	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/water"
	"be-simpletracker/internal/core/tracking/weight"

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
		&steps.StepLog{},
		&weight.BodyWeightLog{},
		&water.WaterLog{},
		&water.DrinkSizePreset{},
	)
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	group := router.Group("/tracking", authMiddleware)
	missed.RegisterMissedRoutes(group, h.db)
	weight.RegisterWeightRoutes(group.Group("/weight"), h.db)
	steps.RegisterStepsRoutes(group.Group("/steps"), h.db)
	water.RegisterWaterRoutes(group.Group("/water"), h.db)
}
