package controller

import (
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetMealPlanToday(c *gin.Context) {
	offset := utils.GetDayOffset(c)
	day, tot, err := services.MealPlanToday(c.Request.Context(), offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
		"today":         time.Now(),
	})
}

func GetMealPlanWeek(c *gin.Context) {
	data, err := services.MealPlanWeek(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"days":  data,
		"today": time.Now(),
	})
}

func GetMonthPlannedSummary(c *gin.Context) {
	offset, err := utils.ParseQueryInt(c, monthOffsetQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	counts, err := services.MonthPlannedSummary(c.Request.Context(), offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"planned_counts": counts,
		"month_offset":   offset,
	})
}

func GetMealPlanMonth(c *gin.Context) {
	offset, err := utils.ParseQueryInt(c, monthOffsetQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	days, startOfMonth, endOfMonth, month, err := services.MealPlanMonth(c.Request.Context(), offset)
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

func GetMealPlanDay(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	day, tot, err := services.MealPlanDay(c.Request.Context(), uint(id64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

func GetGoalsToday(c *gin.Context) {
	goals, err := services.GoalsToday()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, goals)
}
