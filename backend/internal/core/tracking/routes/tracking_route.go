package routes

import (
	"net/http"
	"strconv"

	"be-simpletracker/internal/core/tracking/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TrackingHandler struct {
	db *gorm.DB
}

func NewTrackingHandler(db *gorm.DB) *TrackingHandler {
	return &TrackingHandler{db: db}
}

func RegisterTrackingRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewTrackingHandler(db)
	group.GET("/missed", h.getMissed)
	w := group.Group("/weight")
	{
		w.GET("", h.getWeights)
		w.POST("", h.postWeight)
	}
	s := group.Group("/steps")
	{
		s.GET("", h.getSteps)
		s.POST("", h.postSteps)
	}
}

func parseLimit(c *gin.Context) int {
	raw := c.Query("limit")
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func (h *TrackingHandler) getWeights(c *gin.Context) {
	limit := parseLimit(c)
	rows, err := services.ListBodyWeights(h.db, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": rows})
}

type postWeightBody struct {
	Date      string  `json:"date"`
	WeightLbs float64 `json:"weight_lbs"`
}

func (h *TrackingHandler) postWeight(c *gin.Context) {
	var body postWeightBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := services.ParseDateString(body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	row, err := services.UpsertBodyWeight(h.db, date, body.WeightLbs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}

func (h *TrackingHandler) getMissed(c *gin.Context) {
	date, missingWeight, missingSteps, err := services.GetMissedYesterday(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"date":   date.Format("2006-01-02"),
		"weight": missingWeight,
		"steps":  missingSteps,
	})
}

func (h *TrackingHandler) getSteps(c *gin.Context) {
	limit := parseLimit(c)
	rows, err := services.ListSteps(h.db, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": rows})
}

type postStepsBody struct {
	Date  string `json:"date"`
	Steps int    `json:"steps"`
}

func (h *TrackingHandler) postSteps(c *gin.Context) {
	var body postStepsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := services.ParseDateString(body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	row, err := services.UpsertSteps(h.db, date, body.Steps)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}
