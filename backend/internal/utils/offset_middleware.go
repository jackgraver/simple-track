package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const ginContextKeyDayOffset = "dayOffset"

// DayOffsetMiddleware parses the "offset" query param (days from today: 0 = today)
// and stores it on the gin context for GetDayOffset.
func DayOffsetMiddleware() gin.HandlerFunc {
	spec := QueryIntVar{
		Key:        "offset",
		Default:    0,
		ErrInvalid: "offset must be an integer",
	}
	return func(c *gin.Context) {
		dayOffset, err := ParseQueryInt(c, spec)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set(ginContextKeyDayOffset, dayOffset)
		c.Next()
	}
}

// GetDayOffset returns the day offset set by DayOffsetMiddleware.
func GetDayOffset(c *gin.Context) int {
	v, _ := c.Get(ginContextKeyDayOffset)
	return v.(int)
}
