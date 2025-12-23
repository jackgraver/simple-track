package routes

import (
	"be-simpletracker/features/workout/models"
	"be-simpletracker/features/workout/services"
	generics "be-simpletracker/generics"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkoutPlanHandler struct {
	db *gorm.DB
}

// NewWorkoutPlanHandler creates a new workout plan handler
func NewWorkoutPlanHandler(db *gorm.DB) *WorkoutPlanHandler {
	return &WorkoutPlanHandler{db: db}
}

// RegisterWorkoutPlanRoutes registers all plan routes under the workout group
func RegisterWorkoutPlanRoutes(group *gin.RouterGroup, db *gorm.DB) {
	// Create default config for WorkoutPlan
	config := generics.DefaultCRUDConfig[models.WorkoutPlan]("/plans", "plan")
	
	// Register basic CRUD routes
	generics.RegisterBasicCRUD(group, db, config)

	h := NewWorkoutPlanHandler(db)

    // TODO: Add custom routes for exercise management:
    // group.POST("/plans/:id/exercises/add", f.addExerciseToPlan)
    // group.DELETE("/plans/:id/exercises/remove", f.removeExerciseFromPlan)

	// Day assignment routes
	plans := group.Group("/plans")
	{
		plans.POST("/:id/assign-day", h.assignPlanToDay)
		plans.DELETE("/:id/assign-day", h.unassignPlanFromDay)
	}
}

type AssignDayRequest struct {
	DayOfWeek int `json:"day_of_week" binding:"required,min=0,max=6"`
}

func (h *WorkoutPlanHandler) assignPlanToDay(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID: " + err.Error()})
		return
	}

	var request AssignDayRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
			"details": "Expected JSON: {\"day_of_week\": <number 0-6>}",
			"example": gin.H{"day_of_week": 1},
		})
		return
	}

	plan, err := services.AssignPlanToDay(h.db, uint(planID), request.DayOfWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func (h *WorkoutPlanHandler) unassignPlanFromDay(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	plan, err := services.UnassignPlanFromDay(h.db, uint(planID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}
