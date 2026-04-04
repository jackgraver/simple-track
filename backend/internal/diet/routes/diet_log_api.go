package routes

import (
	"be-simpletracker/internal/diet/services"
	"be-simpletracker/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DietLogHandler struct {
	svc *services.DietLogService
}

func NewDietLogHandler(db *gorm.DB) *DietLogHandler {
	return &DietLogHandler{svc: services.NewDietLogService(db)}
}

func RegisterDietLogRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewDietLogHandler(db)
	logs := group.Group("/logs")
	{
		logs.GET("/today", h.getMealPlanToday)
		logs.GET("/week", h.getMealPlanWeek)
		logs.GET("/month", h.getMealPlanMonth)
		logs.GET("/day/:id", h.getMealPlanDay)
		logs.GET("/goals/today", h.getGoalsToday)
	}
}

func (h *DietLogHandler) getMealPlanToday(c *gin.Context) {
	offset, err := utils.ParseQueryInt(c, weekOffsetQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	day, totalCalories, totalProtein, totalFiber, totalCarbs, err := h.svc.MealPlanToday(c.Request.Context(), offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": totalCalories,
		"totalProtein":  totalProtein,
		"totalFiber":    totalFiber,
		"totalCarbs":    totalCarbs,
		"today":         time.Now(),
	})
}

func (h *DietLogHandler) getMealPlanWeek(c *gin.Context) {
	data, err := h.svc.MealPlanWeek(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"days":  data,
		"today": time.Now(),
	})
}

func (h *DietLogHandler) getMealPlanMonth(c *gin.Context) {
	offset, err := utils.ParseQueryInt(c, monthOffsetQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	days, startOfMonth, endOfMonth, month, err := h.svc.MealPlanMonth(c.Request.Context(), offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"days":  days,
		"today": time.Now(),
		"range": gin.H{
			"start": startOfMonth,
			"end":   endOfMonth,
		},
		"month":  month,
		"offset": offset,
	})
}

func (h *DietLogHandler) getMealPlanDay(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	day, totalCalories, totalProtein, totalFiber, totalCarbs, err := h.svc.MealPlanDay(c.Request.Context(), uint(id64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": totalCalories,
		"totalProtein":  totalProtein,
		"totalFiber":    totalFiber,
		"totalCarbs":    totalCarbs,
	})
}

func (h *DietLogHandler) getGoalsToday(c *gin.Context) {
	goals, err := h.svc.GoalsToday()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, goals)
}
