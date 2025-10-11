package workout

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var workoutDB *gorm.DB

func SetEndpoints(router *gin.Engine, db *gorm.DB) {
    workoutDB = db

    group := router.Group("/workout")
    group.GET("/today", getWorkoutToday)
    group.GET("/all", getWorkoutAll)
}

func getWorkoutToday(c *gin.Context) {
    day, err := GetToday(workoutDB)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, day)
}

func getWorkoutAll(c *gin.Context) {
    days, err := GetAll(workoutDB)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, days)
}   
