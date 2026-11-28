package workoutrepo

import (
	"errors"
	"fmt"
)

// ValidationError is bad input found below the handler (a day outside 0-6, an exercise list that
// doesn't match the plan). Handlers answer it with a 400 and its message.
type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }

// Invalid builds a ValidationError.
func Invalid(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}

// ErrForbidden means the caller can see a shared exercise but may not change it (not its creator
// or the admin), or tried an admin-only action. Handlers answer 403: catalog rows are visible to
// everyone, so 403 leaks nothing.
var ErrForbidden = errors.New("forbidden")

// ErrNotInPlan is returned when removing an exercise the caller's plan doesn't contain. (A plan
// that is missing or someone else's is ErrNotFound, "not found", like everywhere else.)
var ErrNotInPlan = errors.New("exercise not in plan")

// ErrExerciseInUse means a shared exercise can't be deleted because someone's plan or log still
// uses it.
var ErrExerciseInUse = errors.New("exercise is in use")
