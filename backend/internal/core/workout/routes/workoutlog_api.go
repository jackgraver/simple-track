package routes

import (
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils"
	"net/http"

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
		logs.POST("/mobility/pre", h.upsertMobilityPre)
		logs.POST("/mobility/post", h.upsertMobilityPost)
	}
}

func (h *WorkoutLogHandler) getWorkoutToday(reqCtx *gin.Context) {
	// 													       offset
	day, err := h.svc.GetOrCreateToday(reqCtx.Request.Context(), 0)
	if err != nil {
		reqCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reqCtx.JSON(http.StatusOK, day)
}

func (h *WorkoutLogHandler) getWorkoutMonth(reqCtx *gin.Context) {
	offset, err := utils.ParseQueryInt(reqCtx, monthOffsetQuery)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.svc.GetMonthWorkoutLogs(reqCtx.Request.Context(), offset)
	if err != nil {
		reqCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reqCtx.JSON(http.StatusOK, data)
}

type upsertCardioRequest struct {
	Minutes int    `json:"minutes" binding:"required,gte=0"`
	Type    string `json:"type"`
	Notes   string `json:"notes"`
}

func (h *WorkoutLogHandler) upsertCardio(reqCtx *gin.Context) {
	offset, err := utils.ParseQueryInt(reqCtx, weekOffsetQuery)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req upsertCardioRequest
	if err := reqCtx.ShouldBindJSON(&req); err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cardio, err := h.svc.UpsertCardio(reqCtx.Request.Context(), offset, req.Minutes, req.Type, req.Notes)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reqCtx.JSON(http.StatusOK, gin.H{"cardio": cardio})
}

func (h *WorkoutLogHandler) getPreviousWorkout(reqCtx *gin.Context) {
	offset, err := utils.ParseQueryInt(reqCtx, weekOffsetQuery)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := h.svc.GetPreviousWorkoutView(reqCtx.Request.Context(), offset)
	if err != nil {
		reqCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reqCtx.JSON(http.StatusOK, payload)
}

type upsertMobilityRequest struct {
	Checked []string `json:"checked"`
}

func (h *WorkoutLogHandler) upsertMobilityPre(reqCtx *gin.Context) {
	offset, err := utils.ParseQueryInt(reqCtx, weekOffsetQuery)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var req upsertMobilityRequest
	if err := reqCtx.ShouldBindJSON(&req); err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view, err := h.svc.UpsertMobilityPre(reqCtx.Request.Context(), offset, req.Checked)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reqCtx.JSON(http.StatusOK, gin.H{"mobility": view})
}

func (h *WorkoutLogHandler) upsertMobilityPost(reqCtx *gin.Context) {
	offset, err := utils.ParseQueryInt(reqCtx, weekOffsetQuery)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var req upsertMobilityRequest
	if err := reqCtx.ShouldBindJSON(&req); err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view, err := h.svc.UpsertMobilityPost(reqCtx.Request.Context(), offset, req.Checked)
	if err != nil {
		reqCtx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reqCtx.JSON(http.StatusOK, gin.H{"mobility": view})
}
