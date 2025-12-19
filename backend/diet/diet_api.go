package workout

import (
	"be-simpletracker/diet/routes"

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
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	group := router.Group("/diet")
	
	// Register plan sub-domain routes
	routes.RegisterDietPlanRoutes(group, db)
	routes.RegisterMealRoutes(group, db)
	routes.RegisterDietLogRoutes(group, db)
}