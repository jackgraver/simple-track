package routes

import (
	"be-simpletracker/internal/features/workout/models"
	"be-simpletracker/internal/features/workout/services"
	generics "be-simpletracker/internal/generics"
	"net/http"
	"strconv"
	"strings"

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

	plans := group.Group("/plans")
	{
		plans.POST("/:id/exercises/add", h.addExerciseToPlan)
		plans.DELETE("/:id/exercises/remove", h.removeExerciseFromPlan)
		plans.POST("/:id/assign-day", h.assignPlanToDay)
		plans.DELETE("/:id/assign-day", h.unassignPlanFromDay)
		plans.PUT("/:id/planned-cardio", h.setPlannedCardio)
	}
}

type PlanExerciseRequest struct {
	ExerciseID uint `json:"exercise_id" binding:"required"`
}

func (h *WorkoutPlanHandler) addExerciseToPlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	var request PlanExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plan models.WorkoutPlan
	if err := h.db.Preload("Exercises").First(&plan, planID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}
	for _, ex := range plan.Exercises {
		if ex.ID == request.ExerciseID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise already in plan"})
			return
		}
	}

	if err := services.AddExerciseToPlan(h.db, uint(planID), request.ExerciseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Preload("Exercises").First(&plan, planID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func (h *WorkoutPlanHandler) removeExerciseFromPlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	var request PlanExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.RemoveExerciseFromPlan(h.db, uint(planID), request.ExerciseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var plan models.WorkoutPlan
	if err := h.db.Preload("Exercises").First(&plan, uint(planID)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
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
			"error":   "Invalid request body: " + err.Error(),
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

type plannedCardioBody struct {
	Type string `json:"type"`
}

func (h *WorkoutPlanHandler) setPlannedCardio(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}
	var body plannedCardioBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Model(&models.WorkoutPlan{}).
		Where("id = ?", planID).
		Update("planned_cardio_type", strings.TrimSpace(body.Type)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var plan models.WorkoutPlan
	if err := h.db.Preload("Exercises").First(&plan, planID).Error; err != nil {
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
