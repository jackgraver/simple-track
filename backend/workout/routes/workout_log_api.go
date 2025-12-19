package routes

import (
	"be-simpletracker/database/services"
	generics "be-simpletracker/generics"
	"be-simpletracker/workout/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkoutLogHandler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewWorkoutLogHandler(db *gorm.DB) *WorkoutLogHandler {
	return &WorkoutLogHandler{db: db}
}

func RegisterWorkoutLogRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewWorkoutLogHandler(db)

	// Only enable GET /all route, disable other CRUD routes
	config := generics.DefaultCRUDConfig[models.WorkoutLog]("/logs", "log")
	config.EnableGetByID = false
	config.EnableCreate = false
	config.EnableUpdate = false
	config.EnableDelete = false
	generics.RegisterBasicCRUD(group, db, config)

	logs := group.Group("/logs")
	{
		logs.GET("/today", h.getWorkoutToday)
		logs.GET("/month", h.getWorkoutMonth)
		logs.GET("/previous", h.getPreviousWorkout)
	}
}	

func (h *WorkoutLogHandler) getWorkoutToday(c *gin.Context) {
    day, err := services.GetToday(h.db, 0)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, day)
}

func (h *WorkoutLogHandler) getWorkoutMonth(c *gin.Context) {
    offsetStr := c.Query("monthoffset")
    offset, _ := strconv.Atoi(offsetStr)

    today := time.Now()
    target := today.AddDate(0, offset, 0)

    startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
    endOfMonth := startOfMonth.AddDate(0, 1, -1)

    start := startOfMonth.AddDate(0, 0, -int(startOfMonth.Weekday()))
    end := endOfMonth.AddDate(0, 0, 7-int(endOfMonth.Weekday()))

    // data, err := services.WorkoutRange(f.db, start, end)
    data, err := generics.GetByDateRange[*models.WorkoutLog](c.Request.Context(), h.db, start, end)
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

func (f *WorkoutLogHandler) getWorkoutAll(c *gin.Context) {
    days, err := services.GetAll(f.db)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, days)
}   

type ExerciseGroup struct {
    Planned         *models.Exercise       `json:"planned,omitempty"`
    Logged          *models.LoggedExercise `json:"logged,omitempty"`
    Previous        *models.LoggedExercise `json:"previous,omitempty"`
}
func (f *WorkoutLogHandler) getPreviousWorkout(c *gin.Context) {
    offsetStr := c.Query("offset")
    offset, _ := strconv.Atoi(offsetStr)

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