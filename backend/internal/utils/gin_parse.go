package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// QueryIntVar describes a query parameter to parse as an integer and the
// user-facing errors to return when validation fails.
type QueryIntVar struct {
	Key      string
	Default  int
	Required bool
	// ErrMissing is returned when Required is true and the query value is absent or empty.
	ErrMissing string
	// ErrInvalid is returned when the value is present but not a valid base-10 integer.
	ErrInvalid string
}

// ParseQueryInt reads and parses c.Query(v.Key) according to v.
// When Required is false and the value is missing or empty, it returns v.Default with a nil error.
func ParseQueryInt(c *gin.Context, v QueryIntVar) (int, error) {
	s := strings.TrimSpace(c.Query(v.Key))
	if s == "" {
		if v.Required {
			if v.ErrMissing != "" {
				return 0, errors.New(v.ErrMissing)
			}
			return 0, fmt.Errorf("%s is required", v.Key)
		}
		return v.Default, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		if v.ErrInvalid != "" {
			return 0, errors.New(v.ErrInvalid)
		}
		return 0, fmt.Errorf("%s must be an integer", v.Key)
	}
	return n, nil
}
