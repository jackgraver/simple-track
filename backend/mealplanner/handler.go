package mealplanner

import (
    "net/http"

    "be-simpletracker/handlers"

    "github.com/gin-gonic/gin"
)

var MealPlanHandler *handlers.Handlers

func SetEndpoints(router *gin.Engine, h *handlers.Handlers) {
    MealPlanHandler = h

    group := router.Group("/mealplan")
    group.GET("/today", getToday)
    group.GET("/week", getWeek)
    group.GET("/foods/all", getAllFoods)
    group.GET("/meals/all", getAllMeals)
}

func getToday(c *gin.Context) {
    data, err := Today(MealPlanHandler.DB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func getWeek(c *gin.Context) {
    data, err := Week(MealPlanHandler.DB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func getAllFoods(c *gin.Context) {
    data, err := AllFoods(MealPlanHandler.DB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func getAllMeals(c *gin.Context) {
    data, err := AllMeals(MealPlanHandler.DB)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}
