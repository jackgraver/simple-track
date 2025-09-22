package workout

import (
    "net/http"

    "be-simpletracker/handlers"

    "github.com/gin-gonic/gin"
)

var WorkoutHandler *handlers.Handlers

func SetEndpoints(router *gin.Engine, h *handlers.Handlers) {
    WorkoutHandler = h

    group := router.Group("/workout")
    group.GET("/today", handleGetToday)
    group.GET("/week", handleGetWeek)
    group.GET("/exercises/all", handleGetAllExercises)
    group.POST("/exercises", handleAddExercise)
}

func handleGetToday(c *gin.Context) {
    day, err := GetToday(WorkoutHandler.DB)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, day)
}

func handleGetWeek(c *gin.Context) {
    days, err := GetWeek(WorkoutHandler.DB)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, days)
}

func handleGetAllExercises(c *gin.Context) {
    items, err := GetAllExercises(WorkoutHandler.DB)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, items)
}

func handleAddExercise(c *gin.Context) {
    var req struct {
        Name string `json:"name" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ex, err := AddExercise(WorkoutHandler.DB, req.Name)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, ex)
}


