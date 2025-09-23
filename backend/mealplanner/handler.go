package mealplanner

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var mealplannerDB *gorm.DB

func SetEndpoints(router *gin.Engine, db *gorm.DB) {
    mealplannerDB = db

    group := router.Group("/mealplan")
    group.GET("/today", getMealPlanToday)
    group.GET("/week", getMealPlanWeek)
    group.GET("/goals/today", getGoalsToday)
    group.GET("/food/all", getAllFoods)
    group.GET("/meal/all", getAllMeals)
    group.POST("/meal/log", logMeal)
}

func getMealPlanToday(c *gin.Context) {
    data, err := MealPlanToday(mealplannerDB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
	})
}

func getMealPlanWeek(c *gin.Context) {
    data, err := MealPlanWeek(mealplannerDB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
	})
}

func getGoalsToday(c *gin.Context) {
    data, err := GoalsToday(mealplannerDB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func getAllFoods(c *gin.Context) {
    data, err := AllFoods(mealplannerDB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func getAllMeals(c *gin.Context) {
    data, err := AllMeals(mealplannerDB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

type CreateDayMealRequest struct {
	MealID uint   `json:"meal_id"`
	Status string `json:"status"`
}

func logMeal(c *gin.Context) {
	var req CreateDayMealRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse date or use today
	dayDate := time.Now().Truncate(24 * time.Hour)

	// Find or create MealPlanDay
    // var day MealPlanDay
    day, derr := FindMealPlanDay(mealplannerDB, dayDate)
    if derr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": derr.Error()})
        return
    }

	// Create DayMeal
	dayMeal := DayMeal{
		MealPlanDayID: day.ID,
		MealID:        req.MealID,
		Status:        req.Status,
	}

	if err := CreateDayMeal(mealplannerDB, dayMeal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dayMeal)
}