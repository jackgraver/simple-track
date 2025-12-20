package routes

import (
	"be-simpletracker/features/workout/models"
	"be-simpletracker/features/workout/services"
	generics "be-simpletracker/generics"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ExercisesHandler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewExercisesHandler(db *gorm.DB) *ExercisesHandler {
	return &ExercisesHandler{db: db}
}

func RegisterExercisesRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewExercisesHandler(db)

    config := generics.DefaultCRUDConfig[models.Exercise]("/exercises", "exercise")
    config.EnableGetByID = false
	config.EnableUpdate = false
	config.EnableDelete = false
    generics.RegisterBasicCRUD(group, db, config)

	exercises := group.Group("/exercises")
	{
		exercises.POST("/log", h.logExercise)
		exercises.POST("/all-logged", h.checkAllLogged)
		// exercises.GET("/all", h.getAllExercises)
		exercises.POST("/add", h.addExerciseToWorkout)
		exercises.DELETE("/remove", h.removeExerciseFromWorkout)
		exercises.GET("/progression/:id", h.getExerciseProgression)
		// exercises.POST("/create", h.createExercise)
	}
}

type LogExerciseRequest struct {
	Log models.LoggedExercise `json:"exercise"`
    Type string `json:"type"`
}

func (h *ExercisesHandler) logExercise(c *gin.Context) {
	var request LogExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
    }

    switch request.Type {
        case "previous":
            request.Log.ID = 0
            for i := range request.Log.Sets {
                request.Log.Sets[i].LoggedExerciseID = 0
                request.Log.Sets[i].ID = 0
            }
            err := services.LogExercise(h.db, &request.Log)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        case "logged":
            err := services.UpdateLoggedExercise(h.db, request.Log)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
    }

	c.JSON(http.StatusOK, gin.H{"exercise": request.Log})
}

func (h *ExercisesHandler) checkAllLogged(c *gin.Context) {
    today, err := services.GetToday(h.db, 0)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }

    if len(today.Exercises) == len(today.WorkoutPlan.Exercises) {
        fmt.Println("All logged!")
        c.JSON(http.StatusOK, gin.H{"all_logged": true})
        return
    }

    fmt.Println("Not all logged!")
    c.JSON(http.StatusOK, gin.H{"all_logged": false})
}

type AddExerciseRequest struct {
    ExerciseID uint `json:"exercise_id"`
}

func (h *ExercisesHandler) addExerciseToWorkout(c *gin.Context) {
    var request AddExerciseRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    today, err := services.GetToday(h.db, 0)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Check if exercise already exists in workout
    for _, ex := range today.Exercises {
        if ex.ExerciseID == request.ExerciseID {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise already in workout"})
            return
        }
    }

    // Create a new logged exercise entry (empty, ready to be logged)
    newExercise := models.LoggedExercise{
        WorkoutLogID: today.ID,
        ExerciseID:   request.ExerciseID,
        Sets:         []models.LoggedSet{},
        Notes:        "",
    }

    err = services.LogExercise(h.db, &newExercise)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Reload the exercise with Exercise relation using the ID that was set by Create
    var createdExercise models.LoggedExercise
    err = h.db.Preload("Exercise").Preload("Sets").Where("id = ?", newExercise.ID).First(&createdExercise).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"exercise": createdExercise})
}

type RemoveExerciseRequest struct {
    ExerciseID uint `json:"exercise_id"`
}

func (h *ExercisesHandler) removeExerciseFromWorkout(c *gin.Context) {
    var request RemoveExerciseRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    today, err := services.GetToday(h.db, 0)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Find and delete the logged exercise
    var loggedExercise models.LoggedExercise
    err = h.db.Where("workout_log_id = ? AND exercise_id = ?", today.ID, request.ExerciseID).First(&loggedExercise).Error
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found in workout"})
        return
    }

    // Delete the logged exercise (sets will be cascade deleted)
    err = h.db.Delete(&loggedExercise).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ExercisesHandler) getExerciseProgression(c *gin.Context) {
    exerciseIDStr := c.Param("id")
    exerciseID, err := strconv.ParseUint(exerciseIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID"})
        return
    }

    progression, err := services.GetExerciseProgression(h.db, uint(exerciseID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"progression": progression})
}

type CreateExerciseRequest struct {
    Name        string `json:"name"`
    RepRollover uint   `json:"rep_rollover"`
}

func (h *ExercisesHandler) createExercise(c *gin.Context) {
    var request CreateExerciseRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if request.Name == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise name is required"})
        return
    }

    if request.RepRollover == 0 {
        request.RepRollover = 10
    }

    exercise, err := services.CreateExercise(h.db, request.Name, request.RepRollover)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}