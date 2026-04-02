package routes

import (
	"be-simpletracker/internal/workout/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkoutLogHandler struct {
	svc *services.WorkoutLogService
}

func NewWorkoutLogHandler(db *gorm.DB) *WorkoutLogHandler {
	return &WorkoutLogHandler{svc: services.NewWorkoutLogService(db)}
}

func RegisterWorkoutLogRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewWorkoutLogHandler(db)
	logs := group.Group("/logs")
	{
		logs.GET("/today", h.getWorkoutToday)
		logs.GET("/month", h.getWorkoutMonth)
		logs.GET("/previous", h.getPreviousWorkout)
		logs.POST("/cardio", h.upsertCardio)
	}
}

func (h *WorkoutLogHandler) getWorkoutToday(c *gin.Context) {
	day, err := h.svc.GetOrCreateToday(c.Request.Context(), 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, day)
}

func (h *WorkoutLogHandler) getWorkoutMonth(c *gin.Context) {
	offsetStr := c.Query("monthoffset")
	offset, _ := strconv.Atoi(offsetStr)
	data, err := h.svc.GetMonthWorkoutLogs(c.Request.Context(), offset)
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

func (h *WorkoutLogHandler) upsertCardio(c *gin.Context) {
	offsetStr := c.Query("offset")
	offset, _ := strconv.Atoi(offsetStr)
	var req upsertCardioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cardio, err := h.svc.UpsertCardio(c.Request.Context(), offset, req.Minutes, req.Type, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cardio": cardio})
}

func (h *WorkoutLogHandler) getPreviousWorkout(c *gin.Context) {
	offsetStr := c.Query("offset")
	offset, _ := strconv.Atoi(offsetStr)
	payload, err := h.svc.GetPreviousWorkoutView(c.Request.Context(), offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payload)
}
