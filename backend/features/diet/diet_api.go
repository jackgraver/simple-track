package diet

import (
	"be-simpletracker/features/diet/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers all workout feature routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/diet")
	
	// Register plan sub-domain routes
	routes.RegisterDietPlanRoutes(group, h.db)
	routes.RegisterMealRoutes(group, h.db)
	routes.RegisterDietLogRoutes(group, h.db)
}