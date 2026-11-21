package controller

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type workoutProgramRequest struct {
	Name string `json:"name" binding:"required"`
}

type createWorkoutPlanRequest struct {
	Name      string `json:"name" binding:"required"`
	DayOfWeek *int   `json:"day_of_week"`
}

func GetAllWorkoutPrograms(c *gin.Context) {
	programs, err := services.GetAllWorkoutPrograms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"programs": programs})
}

func CreateWorkoutProgram(c *gin.Context) {
	var body workoutProgramRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	program, err := services.CreateWorkoutProgram(body.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"program": program})
}

func CreateWorkoutPlan(c *gin.Context) {
	programID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}
	var body createWorkoutPlanRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := services.CreateWorkoutPlan(uint(programID), body.Name, body.DayOfWeek)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"plan": plan})
}

func RenameWorkoutProgram(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}
	var body workoutProgramRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	program, err := services.RenameWorkoutProgram(uint(id), body.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"program": program})
}

func ActivateWorkoutProgram(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}
	program, err := services.ActivateWorkoutProgram(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"program": program})
}

func GetAllWorkoutPlans(c *gin.Context) {
	workoutPlans, err := services.GetAllWorkoutPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": workoutPlans})
}

type PlanExerciseRequest struct {
	ExerciseID uint `json:"exercise_id" binding:"required"`
}

func AddExerciseToPlan(c *gin.Context) {
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

	plan, err := services.LoadPlanWithOrderedExercises(uint(planID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}
	for _, ex := range plan.Exercises {
		if ex.ID == request.ExerciseID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise already in plan"})
			return
		}
	}

	if err := services.AddExerciseToPlan(uint(planID), request.ExerciseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	plan, err = services.LoadPlanWithOrderedExercises(uint(planID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func RemoveExerciseFromPlan(c *gin.Context) {
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

	if err := services.RemoveExerciseFromPlan(uint(planID), request.ExerciseID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not in plan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	plan, err := services.LoadPlanWithOrderedExercises(uint(planID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

type reorderExercisesBody struct {
	ExerciseIDs []uint `json:"exercise_ids" binding:"required"`
}

func ReorderPlanExercises(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}
	var body reorderExercisesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := services.ReorderPlanExercises(uint(planID), body.ExerciseIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := services.LoadPlanWithOrderedExercises(uint(planID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

type AssignDayRequest struct {
	DayOfWeek *int `json:"day_of_week" binding:"required,gte=0,lte=6"`
}

func AssignPlanToDay(c *gin.Context) {
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

	plan, err := services.AssignPlanToDay(uint(planID), *request.DayOfWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

type plannedCardioBody struct {
	Type    string `json:"type"`
	Minutes int    `json:"minutes"`
}

func SetPlannedCardio(c *gin.Context) {
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
	plan, err := services.SetPlannedCardio(uint(planID), body.Type, body.Minutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

type plannedMobilityBody struct {
	PreMobilityItems  []string `json:"pre_mobility_items"`
	PostMobilityItems []string `json:"post_mobility_items"`
}

func SetPlannedMobility(c *gin.Context) {
	planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}
	var body plannedMobilityBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := services.SetPlannedMobility(uint(planID), body.PreMobilityItems, body.PostMobilityItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func UnassignPlanFromDay(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	var plan *models.WorkoutPlan
	if rawDay := c.Query("day_of_week"); rawDay != "" {
		day, parseErr := strconv.Atoi(rawDay)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day_of_week"})
			return
		}
		plan, err = services.UnassignPlanFromSpecificDay(uint(planID), day)
	} else {
		plan, err = services.UnassignPlanFromDay(uint(planID))
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}
