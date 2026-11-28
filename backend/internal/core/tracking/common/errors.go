package common

import (
	"errors"

	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
)

// InvalidError is a mistake the client made (a missing field, a number out of range). Its message
// is written by us and safe to show; nothing else from a service ever is.
type InvalidError struct{ Msg string }

func (e *InvalidError) Error() string { return e.Msg }

// Invalid returns an InvalidError (answered as 400 by Respond).
func Invalid(msg string) error { return &InvalidError{Msg: msg} }

// Respond answers an error from a tracking service: 400 with its own message for an InvalidError,
// 404 {"error":"not found"} for a row that is missing or someone else's, otherwise a logged 500
// with a fixed message. Database error text never reaches the client (BACKLOG #6).
func Respond(c *gin.Context, err error) {
	var invalid *InvalidError
	if errors.As(err, &invalid) {
		apierr.BadRequest(c, invalid.Msg)
		return
	}
	apierr.NotFoundOrInternal(c, err)
}
