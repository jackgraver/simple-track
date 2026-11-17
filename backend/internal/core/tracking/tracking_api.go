package tracking

import (
	"be-simpletracker/internal/core/tracking/grocery"
	"be-simpletracker/internal/core/tracking/missed"
	"be-simpletracker/internal/core/tracking/profile"
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

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	group := router.Group("/tracking", authMiddleware)
	missed.RegisterMissedRoutes(group, h.db)
	profile.RegisterProfileRoutes(group.Group("/profile"), h.db)
	grocery.RegisterGroceryRoutes(group.Group("/grocery"), h.db)
	weight.RegisterWeightRoutes(group.Group("/weight"), h.db)
	steps.RegisterStepsRoutes(group.Group("/steps"), h.db)
	water.RegisterWaterRoutes(group.Group("/water"), h.db)
}
