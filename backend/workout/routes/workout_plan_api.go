package routes

import (
	generics "be-simpletracker/generics"
	"be-simpletracker/workout/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterWorkoutPlanRoutes registers all plan routes under the workout group
func RegisterWorkoutPlanRoutes(group *gin.RouterGroup, db *gorm.DB) {
	// Create default config for WorkoutPlan
	config := generics.DefaultCRUDConfig[models.WorkoutPlan]("/plans", "plan")
	
	// Register basic CRUD routes
	generics.RegisterBasicCRUD(group, db, config)

	// TODO: Add custom routes for exercise management:
	// group.POST("/plans/:id/exercises/add", f.addExerciseToPlan)
	// group.DELETE("/plans/:id/exercises/remove", f.removeExerciseFromPlan)
}
