package controller

import (
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/utils/apierr"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Per-set endpoints for optimistic logging (optimistic-ui G1). They replace the full-replace
// POST /workout/exercises/log for sets; that endpoint stays for the current frontend and notes.

type logSetRequest struct {
	ClientID string `json:"client_id"`
	// Date is the workout day as YYYY-MM-DD (server local time). Required: resolving the day
	// from ?offset= at request time would put a replay that crosses midnight on a different
	// workout log, which then fails as a client_id conflict instead of returning the stored set.
	Date        string  `json:"date"`
	ExerciseID  uint    `json:"exercise_id"`
	Reps        uint    `json:"reps"`
	Weight      float32 `json:"weight"`
	WeightSetup string  `json:"weight_setup"`
	// Side is "left" or "right" for one side of a unilateral exercise; omitted or null for bilateral.
	Side *string `json:"side"`
	// PairID (a uuid, with a side) ties this side to the other side of the same set number. The
	// first side sends its own client_id; the second sends the first side's.
	PairID *string `json:"pair_id"`
}

// LogSet handles POST /workout/sets. Idempotent by client_id: a replay returns the stored set
// with created=false. Responds with the full logged exercise so the client can reconcile.
func LogSet(c *gin.Context) {
	var req logSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	clientID, ok := normalizeClientID(req.ClientID)
	if !ok {
		apierr.BadRequest(c, "client_id must be a uuid")
		return
	}
	if req.ExerciseID == 0 {
		apierr.BadRequest(c, "exercise_id is required")
		return
	}
	if req.Date == "" {
		apierr.BadRequest(c, "date is required (YYYY-MM-DD)")
		return
	}
	day, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		apierr.BadRequest(c, "date must be YYYY-MM-DD")
		return
	}
	side, ok := models.NormalizeSide(req.Side)
	if !ok {
		apierr.BadRequest(c, "side must be left, right or null")
		return
	}
	var pairID *string
	if req.PairID != nil && *req.PairID != "" && side != nil {
		id, ok := normalizeClientID(*req.PairID)
		if !ok {
			apierr.BadRequest(c, "pair_id must be a uuid")
			return
		}
		pairID = &id
	}

	res, err := services.LogSet(c.Request.Context(), services.LogSetInput{
		ClientID:    clientID,
		Day:         day,
		ExerciseID:  req.ExerciseID,
		Reps:        req.Reps,
		Weight:      req.Weight,
		WeightSetup: req.WeightSetup,
		Side:        side,
		PairID:      pairID,
	})
	switch {
	case errors.Is(err, services.ErrExerciseNotFound):
		apierr.NotFound(c, "Exercise not found")
		return
	case errors.Is(err, workoutrepo.ErrClientIDConflict):
		apierr.Conflict(c, "client_id already used for a different set")
		return
	case err != nil:
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"exercise":         res.Exercise,
		"set":              res.Set,
		"created":          res.Created,
		"first_set_of_day": res.FirstSetOfDay,
	})
}

type updateSetRequest struct {
	Reps        *uint    `json:"reps"`
	Weight      *float32 `json:"weight"`
	WeightSetup *string  `json:"weight_setup"`
	// Side, when present, moves the set to that side ("left"/"right"); "" makes it bilateral.
	Side *string `json:"side"`
	// PairID, when present, moves a side into another set number's pair (a uuid); "" clears it.
	PairID *string `json:"pair_id"`
}

// UpdateSet handles PATCH /workout/sets/:id, where :id is the numeric set id or its client_id.
// Only fields present in the body change, so a replay is harmless.
func UpdateSet(c *gin.Context) {
	ref, ok := parseSetRef(c.Param("id"))
	if !ok {
		apierr.BadRequest(c, "id must be a set id or client_id")
		return
	}
	var req updateSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "Invalid request body")
		return
	}
	if req.Side != nil {
		if _, ok := models.NormalizeSide(req.Side); !ok {
			apierr.BadRequest(c, "side must be left, right or empty")
			return
		}
	}
	if req.PairID != nil && *req.PairID != "" {
		id, ok := normalizeClientID(*req.PairID)
		if !ok {
			apierr.BadRequest(c, "pair_id must be a uuid or empty")
			return
		}
		req.PairID = &id
	}
	set, exercise, err := services.UpdateSet(c.Request.Context(), ref, workoutrepo.SetPatch{
		Reps:        req.Reps,
		Weight:      req.Weight,
		WeightSetup: req.WeightSetup,
		Side:        req.Side,
		PairID:      req.PairID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "Set not found")
			return
		}
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"exercise": exercise, "set": set})
}

// DeleteSet handles DELETE /workout/sets/:id (numeric id or client_id). A set that is already
// gone returns 200 with deleted=false, so retries and double taps are safe.
func DeleteSet(c *gin.Context) {
	ref, ok := parseSetRef(c.Param("id"))
	if !ok {
		apierr.BadRequest(c, "id must be a set id or client_id")
		return
	}
	res, err := services.DeleteSet(c.Request.Context(), ref)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"deleted":            res.Deleted,
		"logged_exercise_id": res.LoggedExerciseID,
		"exercise_removed":   res.ExerciseRemoved,
	})
}

// normalizeClientID validates a uuid and returns it in canonical lowercase form, so the same id
// compares equal on Postgres (uuid column) and SQLite (text).
func normalizeClientID(s string) (string, bool) {
	id, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil || id == uuid.Nil {
		return "", false
	}
	return id.String(), true
}

func parseSetRef(param string) (workoutrepo.SetRef, bool) {
	if n, err := strconv.ParseUint(param, 10, 32); err == nil && n > 0 {
		return workoutrepo.SetRef{ID: uint(n)}, true
	}
	if id, ok := normalizeClientID(param); ok {
		return workoutrepo.SetRef{ClientID: id}, true
	}
	return workoutrepo.SetRef{}, false
}
