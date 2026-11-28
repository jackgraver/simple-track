package service

import (
	"context"
	"errors"
	"fmt"

	"be-simpletracker/internal/database"
)

// ValidationError is returned for bad input and maps to HTTP 400.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

func invalid(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}

// ErrForbidden means the caller can see the row but may not change it (a shared contribution
// type they didn't create, or an admin-only delete). It maps to HTTP 403.
var ErrForbidden = errors.New("forbidden")

// caller is the signed-in user; without one, shared-catalog checks can't be made.
func caller(ctx context.Context) (database.Identity, error) {
	id, ok := database.CurrentIdentity(ctx)
	if !ok {
		return database.Identity{}, database.ErrNoIdentity
	}
	return id, nil
}

func validYear(year int) bool {
	return year >= 1900 && year <= 9999
}
