package controller

import (
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils"
	"be-simpletracker/internal/utils/apierr"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllExercises(c *gin.Context) {
	page, err := utils.ParseQueryInt(c, utils.QueryIntVar{
		Key:        "page",
		Default:    1,
		ErrInvalid: "page must be an integer",
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pageSize, err := utils.ParseQueryInt(c, utils.QueryIntVar{
		Key:        "page_size",
		Default:    0,
		ErrInvalid: "page_size must be an integer",
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	search := c.Query("search")

	result, err := services.ListExercises(page, pageSize, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasNext := false
	if pageSize > 0 {
		hasNext = int64(page*pageSize) < result.Total
	}

	c.JSON(http.StatusOK, gin.H{
		"exercises": result.Exercises,
		"total":     result.Total,
		"has_next":  hasNext,
		"page":      page,
		"page_size": pageSize,
	})
}

type LogExerciseRequest struct {
	Log  models.LoggedExercise `json:"exercise"`
	Type string                `json:"type"`
}

func LogExercise(c *gin.Context) {
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
		err := services.LogExercise(&request.Log)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "logged":
		err := services.UpdateLoggedExercise(request.Log)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": `type must be "previous" or "logged"`})
		return
	}

	savedExercise, err := services.LoadLoggedExercise(request.Log.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload exercise: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": savedExercise})
}

type AddExerciseRequest struct {
	ExerciseID uint `json:"exercise_id"`
}

func getOrCreateTodayOrAbort(c *gin.Context) (models.WorkoutLog, bool) {
	today, err := services.GetOrCreateToday(c.Request.Context(), utils.GetDayOffset(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return models.WorkoutLog{}, false
	}
	return today, true
}

func AddExerciseToWorkout(c *gin.Context) {
	var request AddExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	today, ok := getOrCreateTodayOrAbort(c)
	if !ok {
		return
	}
	for _, ex := range today.Exercises {
		if ex.ExerciseID == request.ExerciseID {
			apierr.Conflict(c, "Exercise already in workout")
			return
		}
	}
	newExercise := models.LoggedExercise{
		WorkoutLogID: today.ID,
		ExerciseID:   request.ExerciseID,
		Sets:         []models.LoggedSet{},
		Notes:        "",
	}
	err := services.LogExercise(&newExercise)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	createdExercise, err := services.LoadLoggedExercise(newExercise.ID)
	if err != nil {
		apierr.Internal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": createdExercise})
}

type RemoveExerciseRequest struct {
	ExerciseID uint `json:"exercise_id"`
}

func RemoveExerciseFromWorkout(c *gin.Context) {
	var request RemoveExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	err := services.RemoveLoggedExerciseForDay(c.Request.Context(), utils.GetDayOffset(c), request.ExerciseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "Exercise not found in workout")
			return
		}
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func DeleteLoggedSet(c *gin.Context) {
	setID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid set ID"})
		return
	}

	err = services.DeleteLoggedSet(uint(setID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Set not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetExerciseProgression(c *gin.Context) {
	exerciseIDStr := c.Param("id")
	exerciseID, err := strconv.ParseUint(exerciseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID"})
		return
	}

	progression, err := services.GetExerciseProgression(uint(exerciseID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if progression == nil {
		progression = []services.ExerciseProgressionEntry{}
	}

	c.JSON(http.StatusOK, gin.H{"progression": progression})
}

type CreateExerciseRequest struct {
	Name        string                  `json:"name"`
	RepRollover uint                    `json:"rep_rollover"`
	Cues        string                  `json:"cues"`
	LoadType    models.ExerciseLoadType `json:"load_type"`
}

func CreateExercise(c *gin.Context) {
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

	exercise, err := services.CreateExercise(request.Name, request.RepRollover, request.Cues, request.LoadType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}

type updateExerciseRequest struct {
	Name        string                  `json:"name"`
	RepRollover uint                    `json:"rep_rollover"`
	Cues        string                  `json:"cues"`
	LoadType    models.ExerciseLoadType `json:"load_type"`
}

func UpdateExercise(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID"})
		return
	}
	var req updateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise name is required"})
		return
	}
	if req.RepRollover == 0 {
		req.RepRollover = 10
	}
	exercise, err := services.UpdateExercise(uint(id64), req.Name, req.RepRollover, req.Cues, req.LoadType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}

type updateExerciseCuesRequest struct {
	Cues string `json:"cues"`
}

func UpdateExerciseCues(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID"})
		return
	}
	var req updateExerciseCuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exercise, err := services.UpdateExerciseCues(uint(id64), req.Cues)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}
