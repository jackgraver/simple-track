package routes

import (
	"be-simpletracker/features/diet/models"
	"be-simpletracker/features/diet/services"
	"be-simpletracker/generics"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DietLogHandler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewDietLogHandler(db *gorm.DB) *DietLogHandler {
	return &DietLogHandler{db: db}
}

func RegisterDietLogRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewDietLogHandler(db)

    config := generics.DefaultCRUDConfig[models.DayLog]("/logs", "log")
    generics.RegisterBasicCRUD(group, db, config)

	logs := group.Group("/logs")
	{
        logs.GET("/today", h.getMealPlanToday)
        logs.GET("/week", h.getMealPlanWeek)
        logs.GET("/month", h.getMealPlanMonth)
        logs.GET("/day/:id" , h.getMealPlanDay)
        logs.GET("/goals/today", h.getGoalsToday)
	}
}

func (h *DietLogHandler) getMealPlanToday(c *gin.Context) {
    offsetStr := c.Query("offset")
    offset, _ := strconv.Atoi(offsetStr)

    day, daysErr := services.MealPlanToday(h.db, offset)
    if daysErr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": daysErr.Error()})
        return
    }
    
    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(h.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
		"day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
		"today": time.Now(),
	})
}

func (h *DietLogHandler) getMealPlanWeek(c *gin.Context) {
    today := time.Now()
    start := today.AddDate(0, 0, -3) // 3 days before
	end := today.AddDate(0, 0, 3)    // 3 days after
    // data, err := services.MealPlanRange(f.db, today, start, end)
    data, err := generics.GetByDateRange[*models.Day](c.Request.Context(), h.db, start, end)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
	})
}

func (h *DietLogHandler) getMealPlanMonth(c *gin.Context) {
    offsetStr := c.Query("monthoffset")
    offset, _ := strconv.Atoi(offsetStr)

    today := time.Now()
    target := today.AddDate(0, offset, 0)

    startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
    endOfMonth := startOfMonth.AddDate(0, 1, -1)

    data, err := generics.GetByDateRange[*models.Day](c.Request.Context(), h.db, startOfMonth, endOfMonth)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
		"range": gin.H{
			"start": startOfMonth,
			"end": endOfMonth,
		},
		"month": target.Month(),
		"offset": offset,
	})
}

func (h *DietLogHandler) getMealPlanDay(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    day, err := generics.GetByID[*models.Day](c.Request.Context(), h.db, uint(id))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(h.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

func (h *DietLogHandler) getGoalsToday(c *gin.Context) {
    goals, err := services.GoalsToday(h.db)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, goals)
}