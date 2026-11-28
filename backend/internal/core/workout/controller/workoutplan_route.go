package controller

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils/apierr"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type workoutProgramRequest struct {
	Name string `json:"name" binding:"required"`
}

type createWorkoutPlanRequest struct {
	Name      string `json:"name" binding:"required"`
	DayOfWeek *int   `json:"day_of_week"`
}

func GetAllWorkoutPrograms(c *gin.Context) {
	programs, err := services.GetAllWorkoutPrograms(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"programs": programs})
}

func CreateWorkoutProgram(c *gin.Context) {
	var body workoutProgramRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		apierr.BadRequest(c, "name is required")
		return
	}
	program, err := services.CreateWorkoutProgram(c.Request.Context(), body.Name)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"program": program})
}

func CreateWorkoutPlan(c *gin.Context) {
	programID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid program ID")
		return
	}
	var body createWorkoutPlanRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	plan, err := services.CreateWorkoutPlan(c.Request.Context(), uint(programID), body.Name, body.DayOfWeek)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"plan": plan})
}

func RenameWorkoutProgram(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid program ID")
		return
	}
	var body workoutProgramRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		apierr.BadRequest(c, "name is required")
		return
	}
	program, err := services.RenameWorkoutProgram(c.Request.Context(), uint(id), body.Name)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"program": program})
}

func ActivateWorkoutProgram(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid program ID")
		return
	}
	program, err := services.ActivateWorkoutProgram(c.Request.Context(), uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"program": program})
}

func GetAllWorkoutPlans(c *gin.Context) {
	workoutPlans, err := services.GetAllWorkoutPlans(c.Request.Context())
	if err != nil {
		respondError(c, err)
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
		apierr.BadRequest(c, "Invalid plan ID")
		return
	}

	var request PlanExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}

	plan, err := services.LoadPlanWithOrderedExercises(c.Request.Context(), uint(planID))
	if err != nil {
		respondError(c, err)
		return
	}
	for _, ex := range plan.Exercises {
		if ex.ID == request.ExerciseID {
			apierr.BadRequest(c, "Exercise already in plan")
			return
		}
	}

	if err := services.AddExerciseToPlan(c.Request.Context(), uint(planID), request.ExerciseID); err != nil {
		respondError(c, err)
		return
	}

	plan, err = services.LoadPlanWithOrderedExercises(c.Request.Context(), uint(planID))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func RemoveExerciseFromPlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid plan ID")
		return
	}

	var request PlanExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}

	if err := services.RemoveExerciseFromPlan(c.Request.Context(), uint(planID), request.ExerciseID); err != nil {
		if errors.Is(err, workoutrepo.ErrNotInPlan) {
			apierr.NotFound(c, "Exercise not in plan")
			return
		}
		respondError(c, err)
		return
	}

	plan, err := services.LoadPlanWithOrderedExercises(c.Request.Context(), uint(planID))
	if err != nil {
		respondError(c, err)
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
		apierr.BadRequest(c, "Invalid plan ID")
		return
	}
	var body reorderExercisesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	if err := services.ReorderPlanExercises(c.Request.Context(), uint(planID), body.ExerciseIDs); err != nil {
		respondError(c, err)
		return
	}
	plan, err := services.LoadPlanWithOrderedExercises(c.Request.Context(), uint(planID))
	if err != nil {
		respondError(c, err)
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
		apierr.BadRequest(c, "Invalid plan ID")
		return
	}

	var request AssignDayRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.BadRequest(c, "day_of_week must be a number from 0 to 6")
		return
	}

	plan, err := services.AssignPlanToDay(c.Request.Context(), uint(planID), *request.DayOfWeek)
	if err != nil {
		respondError(c, err)
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
		apierr.BadRequest(c, "Invalid plan ID")
		return
	}
	var body plannedCardioBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	plan, err := services.SetPlannedCardio(c.Request.Context(), uint(planID), body.Type, body.Minutes)
	if err != nil {
		respondError(c, err)
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
		apierr.BadRequest(c, "Invalid plan ID")
		return
	}
	var body plannedMobilityBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	plan, err := services.SetPlannedMobility(c.Request.Context(), uint(planID), body.PreMobilityItems, body.PostMobilityItems)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func UnassignPlanFromDay(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid plan ID")
		return
	}

	var plan *models.WorkoutPlan
	if rawDay := c.Query("day_of_week"); rawDay != "" {
		day, parseErr := strconv.Atoi(rawDay)
		if parseErr != nil {
			apierr.BadRequest(c, "Invalid day_of_week")
			return
		}
		plan, err = services.UnassignPlanFromSpecificDay(c.Request.Context(), uint(planID), day)
	} else {
		plan, err = services.UnassignPlanFromDay(c.Request.Context(), uint(planID))
	}
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func parsePlanID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid plan ID")
		return 0, false
	}
	return uint(id), true
}

func RenameWorkoutPlan(c *gin.Context) {
	id, ok := parsePlanID(c)
	if !ok {
		return
	}
	var body workoutProgramRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		apierr.BadRequest(c, "name is required")
		return
	}
	plan, err := services.RenameWorkoutPlan(c.Request.Context(), id, body.Name)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

func DeleteWorkoutPlan(c *gin.Context) {
	id, ok := parsePlanID(c)
	if !ok {
		return
	}
	if err := services.DeleteWorkoutPlan(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func RestoreWorkoutPlan(c *gin.Context) {
	id, ok := parsePlanID(c)
	if !ok {
		return
	}
	plan, err := services.RestoreWorkoutPlan(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}
