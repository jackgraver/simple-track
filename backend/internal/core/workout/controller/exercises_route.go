package controller

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
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
		respondError(c, err)
		return
	}
	pageSize, err := utils.ParseQueryInt(c, utils.QueryIntVar{
		Key:        "page_size",
		Default:    0,
		ErrInvalid: "page_size must be an integer",
	})
	if err != nil {
		respondError(c, err)
		return
	}
	search := c.Query("search")

	result, err := services.ListExercises(c.Request.Context(), page, pageSize, search)
	if err != nil {
		respondError(c, err)
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

// logExerciseRequest is the legacy full-replace body. It is a DTO, not models.LoggedExercise, so
// nothing but these fields is ever read from the client: no user_id, no nested exercise, no
// client_id (BACKLOG #1).
type logExerciseRequest struct {
	Log  loggedExerciseBody `json:"exercise"`
	Type string             `json:"type"`
}

type loggedExerciseBody struct {
	ID           uint            `json:"ID"`
	WorkoutLogID uint            `json:"workout_log_id"`
	ExerciseID   uint            `json:"exercise_id"`
	Notes        string          `json:"notes"`
	Sets         []loggedSetBody `json:"sets"`
}

type loggedSetBody struct {
	ID          uint    `json:"ID"`
	Reps        uint    `json:"reps"`
	Weight      float32 `json:"weight"`
	WeightSetup string  `json:"weight_setup"`
	Side        *string `json:"side"`
	PairID      *string `json:"pair_id"`
}

func (b loggedExerciseBody) model() models.LoggedExercise {
	out := models.LoggedExercise{WorkoutLogID: b.WorkoutLogID, ExerciseID: b.ExerciseID, Notes: b.Notes}
	out.ID = b.ID
	for _, s := range b.Sets {
		set := models.LoggedSet{Reps: s.Reps, Weight: s.Weight, WeightSetup: s.WeightSetup, Side: s.Side, PairID: s.PairID}
		set.ID = s.ID
		out.Sets = append(out.Sets, set)
	}
	return out
}

func LogExercise(c *gin.Context) {
	var request logExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	// client_id belongs to the per-set endpoints (POST /workout/sets); the DTO has no field for it.
	log := request.Log.model()

	switch request.Type {
	case "previous":
		log.ID = 0
		for i := range log.Sets {
			log.Sets[i].LoggedExerciseID = 0
			log.Sets[i].ID = 0
		}
		err := services.LogExercise(c.Request.Context(), &log)
		if errors.Is(err, workoutrepo.ErrLoggedExerciseExists) {
			// A double tap or retry already created the row (log.ID now points at it).
			// Treat this request as the full replace it describes instead of adding a duplicate.
			err = services.UpdateLoggedExercise(c.Request.Context(), log)
		}
		if err != nil {
			respondError(c, err)
			return
		}
	case "logged":
		err := services.UpdateLoggedExercise(c.Request.Context(), log)
		if err != nil {
			respondError(c, err)
			return
		}
	default:
		apierr.BadRequest(c, `type must be "previous" or "logged"`)
		return
	}

	savedExercise, err := services.LoadLoggedExercise(c.Request.Context(), log.ID)
	if err != nil {
		respondError(c, err)
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
		respondError(c, err)
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
	err := services.LogExercise(c.Request.Context(), &newExercise)
	if errors.Is(err, workoutrepo.ErrLoggedExerciseExists) {
		apierr.Conflict(c, "Exercise already in workout")
		return
	}
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	createdExercise, err := services.LoadLoggedExercise(c.Request.Context(), newExercise.ID)
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
		apierr.BadRequest(c, "Invalid set ID")
		return
	}

	err = services.DeleteLoggedSet(c.Request.Context(), uint(setID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "not found")
			return
		}
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetExerciseProgression(c *gin.Context) {
	exerciseIDStr := c.Param("id")
	exerciseID, err := strconv.ParseUint(exerciseIDStr, 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid exercise ID")
		return
	}

	progression, err := services.GetExerciseProgression(c.Request.Context(), uint(exerciseID))
	if err != nil {
		respondError(c, err)
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
		apierr.BadRequest(c, "Invalid request body")
		return
	}

	if request.Name == "" {
		apierr.BadRequest(c, "Exercise name is required")
		return
	}

	if request.RepRollover == 0 {
		request.RepRollover = 10
	}

	exercise, err := services.CreateExercise(c.Request.Context(), request.Name, request.RepRollover, request.Cues, request.LoadType)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}

type updateExerciseRequest struct {
	Name        string                  `json:"name"`
	RepRollover uint                    `json:"rep_rollover"`
	Cues        string                  `json:"cues"`
	LoadType    models.ExerciseLoadType `json:"load_type"`
	// Unilateral, when present, turns one-side-at-a-time logging on or off.
	Unilateral *bool `json:"unilateral"`
}

func UpdateExercise(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid exercise ID")
		return
	}
	var req updateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	if req.Name == "" {
		apierr.BadRequest(c, "Exercise name is required")
		return
	}
	if req.RepRollover == 0 {
		req.RepRollover = 10
	}
	exercise, err := services.UpdateExerciseWithUnilateral(c.Request.Context(), uint(id64), req.Name, req.RepRollover, req.Cues, req.Unilateral, req.LoadType)
	if err != nil {
		respondError(c, err)
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
		apierr.BadRequest(c, "Invalid exercise ID")
		return
	}
	var req updateExerciseCuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	exercise, err := services.UpdateExerciseCues(c.Request.Context(), uint(id64), req.Cues)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}

type machineSettingBody struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type exerciseSettingsRequest struct {
	MachineSettings []machineSettingBody `json:"machine_settings"`
}

// SaveExerciseSettings handles PUT /workout/exercises/:id/settings: the caller's own machine
// settings for a shared exercise. Any signed-in user may save theirs for any catalog exercise; an
// empty list clears them.
func SaveExerciseSettings(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid exercise ID")
		return
	}
	var req exerciseSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	settings := make([]models.MachineSetting, 0, len(req.MachineSettings))
	for _, m := range req.MachineSettings {
		settings = append(settings, models.MachineSetting{Label: m.Label, Value: m.Value})
	}
	exercise, err := services.SaveMachineSettings(c.Request.Context(), uint(id64), settings)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}

// DeleteExercise handles DELETE /workout/exercises/:id. Admin only (403 for everyone else); API
// only, there is no UI for it yet.
func DeleteExercise(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid exercise ID")
		return
	}
	if err := services.DeleteExercise(c.Request.Context(), uint(id64)); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
