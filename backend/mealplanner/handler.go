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
}

func getMealPlanToday(c *gin.Context) {
    data, err := MealPlanToday(mealplannerDB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
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