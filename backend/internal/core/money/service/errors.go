package service

import "fmt"

// ValidationError is returned for bad input and maps to HTTP 400.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

func invalid(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}

func validYear(year int) bool {
	return year >= 1900 && year <= 9999
}
