package controller

import (
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/utils/apierr"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// respondError answers a service error: validation 400 with its message, 403 for a shared
// exercise the caller may not change, 404 for a row that is missing or not the caller's (same
// body for both), 409 for a unique-index race or an exercise still in use, otherwise a logged 500
// with a fixed message. Database error text never reaches the client.
func respondError(c *gin.Context, err error) {
	var invalid services.ValidationError
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &invalid):
		apierr.BadRequest(c, invalid.Message)
	case errors.Is(err, services.ErrForbidden):
		apierr.Forbidden(c, "You can't change this exercise")
	case errors.Is(err, services.ErrExerciseInUse):
		apierr.Conflict(c, "Exercise is still used by a routine or workout")
	case errors.Is(err, services.ErrExerciseNotFound):
		apierr.NotFound(c, "Exercise not found")
	case database.IsNotFound(err):
		apierr.NotFound(c, "not found")
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		apierr.Conflict(c, "Already exists")
	default:
		apierr.Internal(c, err)
	}
}
