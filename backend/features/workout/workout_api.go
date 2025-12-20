package workout

import (
	"be-simpletracker/features/workout/routes"

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
	group := router.Group("/workout")
	
	// Register plan sub-domain routes
	routes.RegisterWorkoutPlanRoutes(group, h.db)
	routes.RegisterExercisesRoutes(group, h.db)	
	routes.RegisterWorkoutLogRoutes(group, h.db)
}