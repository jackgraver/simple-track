package routes

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

type ExercisesHandler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewExercisesHandler(db *gorm.DB) *ExercisesHandler {
	return &ExercisesHandler{db: db}
}

func RegisterExercisesRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewExercisesHandler(db)
	dayOffsetMiddleware := utils.DayOffsetMiddleware()
	exercises := group.Group("/exercises")
	{
		exercises.GET("/all", h.getAllExercises)
		exercises.POST("", h.createExercise)
		exercises.PUT("/:id", h.updateExercise)
		exercises.PUT("/:id/cues", h.updateExerciseCues)
		exercises.POST("/log", h.logExercise)
		exercises.POST("/add", dayOffsetMiddleware, h.addExerciseToWorkout)
		exercises.DELETE("/remove", dayOffsetMiddleware, h.removeExerciseFromWorkout)
		exercises.DELETE("/sets/:id", h.deleteLoggedSet)
		exercises.GET("/progression/:id", h.getExerciseProgression)
	}
}

func (h *ExercisesHandler) getAllExercises(c *gin.Context) {
	page, err := utils.ParseQueryInt(c, pageQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pageSize, err := utils.ParseQueryInt(c, pageSizeQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	search := c.Query("search")

	query := h.db.Model(&models.Exercise{})
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	query = query.Order("name ASC")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var exercises []models.Exercise
	if pageSize > 0 {
		query = query.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	if err := query.Find(&exercises).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasNext := false
	if pageSize > 0 {
		hasNext = int64(page*pageSize) < total
	}

	c.JSON(http.StatusOK, gin.H{
		"exercises": exercises,
		"total":     total,
		"has_next":  hasNext,
		"page":      page,
		"page_size": pageSize,
	})
}

type LogExerciseRequest struct {
	Log  models.LoggedExercise `json:"exercise"`
	Type string                `json:"type"`
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
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": `type must be "previous" or "logged"`})
		return
	}

	// Reload the exercise with all relations to get updated IDs
	var savedExercise models.LoggedExercise
	err := h.db.Preload("Exercise").Preload("Sets").Where("id = ?", request.Log.ID).First(&savedExercise).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload exercise: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": savedExercise})
}

type AddExerciseRequest struct {
	ExerciseID uint `json:"exercise_id"`
}

func getOrCreateTodayOrAbort(c *gin.Context, db *gorm.DB) (models.WorkoutLog, bool) {
	today, err := services.GetOrCreateToday(c.Request.Context(), db, utils.GetDayOffset(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return models.WorkoutLog{}, false
	}
	return today, true
}

func (h *ExercisesHandler) addExerciseToWorkout(c *gin.Context) {
	var request AddExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	today, ok := getOrCreateTodayOrAbort(c, h.db)
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
	err := services.LogExercise(h.db, &newExercise)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	var createdExercise models.LoggedExercise
	err = h.db.Preload("Exercise").Preload("Sets").Where("id = ?", newExercise.ID).First(&createdExercise).Error
	if err != nil {
		apierr.Internal(c, err)
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
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	err := services.RemoveLoggedExerciseForDay(c.Request.Context(), h.db, utils.GetDayOffset(c), request.ExerciseID)
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

func (h *ExercisesHandler) deleteLoggedSet(c *gin.Context) {
	setID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid set ID"})
		return
	}

	err = services.DeleteLoggedSet(h.db, uint(setID))
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

	if progression == nil {
		progression = []services.ExerciseProgressionEntry{}
	}

	c.JSON(http.StatusOK, gin.H{"progression": progression})
}

type CreateExerciseRequest struct {
	Name        string `json:"name"`
	RepRollover uint   `json:"rep_rollover"`
	Cues        string `json:"cues"`
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

	exercise, err := services.CreateExercise(h.db, request.Name, request.RepRollover, request.Cues)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}

type updateExerciseRequest struct {
	Name        string `json:"name"`
	RepRollover uint   `json:"rep_rollover"`
	Cues        string `json:"cues"`
}

func (h *ExercisesHandler) updateExercise(c *gin.Context) {
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
	exercise, err := services.UpdateExercise(h.db, uint(id64), req.Name, req.RepRollover, req.Cues)
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

func (h *ExercisesHandler) updateExerciseCues(c *gin.Context) {
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
	exercise, err := services.UpdateExerciseCues(h.db, uint(id64), req.Cues)
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