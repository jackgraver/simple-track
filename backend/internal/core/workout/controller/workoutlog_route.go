package controller

import (
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetWorkoutToday(c *gin.Context) {
	day, err := services.GetOrCreateToday(c.Request.Context(), utils.GetDayOffset(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, day)
}

func GetWorkoutMonth(c *gin.Context) {
	offset, err := utils.ParseQueryInt(c, monthOffsetQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := services.GetMonthWorkoutLogs(c.Request.Context(), offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

type upsertCardioRequest struct {
	Minutes int    `json:"minutes" binding:"required,gte=0"`
	Type    string `json:"type"`
	Notes   string `json:"notes"`
}

func UpsertCardio(c *gin.Context) {
	offset := utils.GetDayOffset(c)
	var req upsertCardioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cardio, err := services.UpsertCardio(c.Request.Context(), offset, req.Minutes, req.Type, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cardio": cardio})
}

func GetPreviousWorkout(c *gin.Context) {
	offset := utils.GetDayOffset(c)
	payload, err := services.GetPreviousWorkoutView(c.Request.Context(), offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payload)
}

type switchPlanRequest struct {
	PlanID *uint `json:"plan_id"`
}

func SwitchPlan(c *gin.Context) {
	offset := utils.GetDayOffset(c)
	var req switchPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload, err := services.SwitchPlan(c.Request.Context(), offset, req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func GetWorkoutActivity(c *gin.Context) {
	mode := strings.TrimSpace(strings.ToLower(c.Query("mode")))
	if mode == "" {
		mode = "rolling"
	}

	weeks, err := utils.ParseQueryInt(c, activityWeeksQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := services.GetWorkoutActivity(c.Request.Context(), mode, weeks)
	if err != nil {
		if errors.Is(err, services.ErrInvalidActivityMode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be year or rolling"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

type upsertMobilityRequest struct {
	Checked []string `json:"checked"`
}

func UpsertMobilityPre(c *gin.Context) {
	offset := utils.GetDayOffset(c)
	var req upsertMobilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view, err := services.UpsertMobilityPre(c.Request.Context(), offset, req.Checked)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mobility": view})
}

func UpsertMobilityPost(c *gin.Context) {
	offset := utils.GetDayOffset(c)
	var req upsertMobilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view, err := services.UpsertMobilityPost(c.Request.Context(), offset, req.Checked)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mobility": view})
}
