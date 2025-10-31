package api

import (
	"be-simpletracker/db/models"
	"be-simpletracker/db/services"
	"fmt"
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
    var feature = models.NewWorkoutModel(db)
    feature.MigrateDatabase()

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
	group.POST("/exercise/log", f.logExercise)
}

func (f *WorkoutFeature) getWorkoutToday(c *gin.Context) {
    day, err := services.GetToday(f.db, 0)
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

func totalVolume(log models.LoggedExercise) float32 {
	var total float32
	for _, s := range log.Sets {
		total += s.Weight * float32(s.Reps)
	}
	return total
}


//TODO: needs a lot more work, I think we need more data before were able to do more complex stuff like this
/*
		prevLog, err := services.GetPreviousExerciseLog(f.db, today.Date, exercise.Name, -1)
		if err != nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
			return
		}

		currentVol := totalVolume(exerciseLog)
		prevVol := totalVolume(prevLog)

		if prevVol > 0 {
			difference := ((currentVol - prevVol) / prevVol) * 100
			if math.Abs(float64(difference)) >= 5 {
				exerciseLog.PercentChange = float32(difference)
			}
		}
*/
type ExerciseGroup struct {
    Planned         *models.Exercise       `json:"planned,omitempty"`
    Logged          *models.LoggedExercise `json:"logged,omitempty"`
    Previous        *models.LoggedExercise `json:"previous,omitempty"`
}
func (f *WorkoutFeature) getPreviousWorkout(c *gin.Context) {
    offsetStr := c.Query("offset")
    offset, _ := strconv.Atoi(offsetStr)
    fmt.Println("offset", offset)

    today, err := services.GetToday(f.db, offset)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }

    logged := today.Exercises
    planned := today.WorkoutPlan.Exercises

    // Build a map for logged exercises keyed by name
    loggedMap := make(map[string]models.LoggedExercise)
    for _, l := range logged {
        if l.Exercise != nil {
            loggedMap[l.Exercise.Name] = l
        }
    }

    // Prepare grouped list
    results := make([]ExerciseGroup, 0)

    // First, handle planned exercises (ensures order by plan)
    for _, p := range planned {
        group := ExerciseGroup{Planned: &p}

        // if already logged, attach it
        if log, ok := loggedMap[p.Name]; ok {
            group.Logged = &log
            delete(loggedMap, p.Name) // remove to avoid duplicates
        }

        // get previous log
        prev, err := services.GetPreviousExerciseLog(f.db, today.Date, p.Name, 0)
        if err == nil {
            group.Previous = &prev
        }

        results = append(results, group)
    }

    // Any leftover logged exercises not part of the plan (edge case)
    for _, l := range loggedMap {
        prev, err := services.GetPreviousExerciseLog(f.db, today.Date, l.Exercise.Name, 0)
        if err == nil {
            results = append(results, ExerciseGroup{
                Logged:   &l,
                Previous: &prev,
            })
        } else {
            results = append(results, ExerciseGroup{Logged: &l})
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "day":        today,
        "previous_exercises":  results,
    })
}


type LogExerciseRequest struct {
	Exercise models.LoggedExercise `json:"exercise"`
}

func (f *WorkoutFeature) logExercise(c *gin.Context) {
	var request LogExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.Exercise.ID = 0

	err := services.LogExercise(f.db, request.Exercise)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": request.Exercise})
}