package workout

import (
	"be-simpletracker/workout/routes"

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
	group := router.Group("/workout")
	
	// Register plan sub-domain routes
	routes.RegisterWorkoutPlanRoutes(group, db)
	routes.RegisterExercisesRoutes(group, db)	
	routes.RegisterWorkoutLogRoutes(group, db)
}