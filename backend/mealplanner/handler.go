package mealplanner

import (
	"fmt"
	"net/http"
	"strconv"
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
    group.GET("/month", getMealPlanMonth)
    group.GET("/day/:id" , getMealPlanDay)
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
    today := time.Now()
    start := today.AddDate(0, 0, -3) // 3 days before
	end := today.AddDate(0, 0, 3)    // 3 days after
    data, err := MealPlanRange(mealplannerDB, today, start, end)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
	})
}

func getMealPlanMonth(c *gin.Context) {
    today := time.Now()
    start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
    end := start.AddDate(0, 1, -1) 
    data, err := MealPlanRange(mealplannerDB, today, start, end)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
	})
}

func getMealPlanDay(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    data, err := MealPlanDayByID(mealplannerDB, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if data == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
        return
    }

    c.JSON(http.StatusOK, data)
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
    time.Sleep(3 * time.Second)
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
    MealID uint       `json:"meal_id"`
    Name   string     `json:"name"`
    Items  []MealItem `json:"items"`
}

func logMeal(c *gin.Context) {
    var req CreateDayMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err) // log to console
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 1. Find or create MealPlanDay
    dayDate := time.Now().Truncate(24 * time.Hour)
    day, derr := FindMealPlanDay(mealplannerDB, dayDate)
    if derr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": derr.Error()})
        return
    }

    var mealID uint

    fmt.Println("req.MealID", req.MealID)
    if req.MealID != 0 {
        fmt.Println("Meal Exists")
        // Meal exists: use it
        mealID = req.MealID
    } else {
        fmt.Println("Meal Doesn't Exist")
        // Meal doesn't exist: create it
        if req.Name == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Meal name is required for new meal"})
            return
        }

        newMeal := Meal{
            Name:  req.Name,
            Items: req.Items,
        }

        if err := mealplannerDB.Create(&newMeal).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        mealID = newMeal.ID
    }

    // 2. Create DayMeal
    dayMeal := DayLog{
        DayID: day.ID,
        MealID:        mealID,
    }

    if err := CreateDayMeal(mealplannerDB, &dayMeal); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Optionally preload Meal and Items for response
    mealplannerDB.Preload("Meal.Items").First(&dayMeal, dayMeal.ID)

    c.JSON(http.StatusOK, dayMeal)
}
