package api

import (
	"be-simpletracker/db/models"
	"be-simpletracker/db/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkoutFeature struct {
	BaseFeature[models.WorkoutModel]
}

func NewWorkoutFeature(db *gorm.DB) *WorkoutFeature {
    models.NewWorkoutModel(db)
    // var feature = models.NewWorkoutModel(db)
    // feature.MigrateDatabase()

	return &WorkoutFeature{
		BaseFeature[models.WorkoutModel]{
			db: db,
		},
	}
}

func (f *WorkoutFeature) SetEndpoints(router *gin.Engine) {
    group := router.Group("/workout")
    group.GET("/today", f.getWorkoutToday)
    group.GET("/month", f.getWorkoutMonth)
    group.GET("/all", f.getWorkoutAll)
    group.GET("/previous", f.getPreviousWorkout)
}

func (f *WorkoutFeature) getWorkoutToday(c *gin.Context) {
    day, err := services.GetToday(f.db)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, day)
}


func (f *WorkoutFeature) getWorkoutMonth(c *gin.Context) {
    offsetStr := c.Query("monthoffset")
    offset, _ := strconv.Atoi(offsetStr)

    today := time.Now()
    target := today.AddDate(0, offset, 0)

    startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
    endOfMonth := startOfMonth.AddDate(0, 1, -1)

    start := startOfMonth.AddDate(0, 0, -int(startOfMonth.Weekday()))
    end := endOfMonth.AddDate(0, 0, 7-int(endOfMonth.Weekday()))

    // data, err := services.WorkoutRange(f.db, start, end)
    data, err := services.ObjectRange[*models.WorkoutLog](f.db, start, end)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "days":  data,
        "today": today,
        "range": gin.H{
            "start": start,
            "end":   end,
        },
        "month": target.Month(),
        "offset": offset,
    })
}

func (f *WorkoutFeature) getWorkoutAll(c *gin.Context) {
    days, err := services.GetAll(f.db)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, days)
}   

func (f *WorkoutFeature) getPreviousWorkout(c *gin.Context) {
    today, err := services.GetToday(f.db)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }

    previousExerciseLogs := make([]models.LoggedExercise, 0)
    for _, exercise := range today.Exercises {
        exerciseLog, err := services.GetPreviousExerciseLog(f.db, today.Date, exercise.Name)
        if err != nil {
            c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
            return
        }
        previousExerciseLogs = append(previousExerciseLogs, exerciseLog)
    }

    c.JSON(http.StatusOK, gin.H{"exercises": previousExerciseLogs, "day": today})
}