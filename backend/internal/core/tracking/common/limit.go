package common

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseLimitQuery reads optional positive integer limit query param; 0 means use service default.
func ParseLimitQuery(c *gin.Context) int {
	raw := c.Query("limit")
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
