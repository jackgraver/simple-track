package utils

import (
	"be-simpletracker/internal/clock"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const ginContextKeyDayOffset = "dayOffset"

// DayOffsetMiddleware parses the "offset" query param (days from today: 0 = today)
// and stores it on the gin context for GetDayOffset. A "date" param (YYYY-MM-DD) takes
// precedence: it is the server-local calendar day, the same way POST /workout/sets reads its
// date, so a client in another time zone reads and writes the same day.
func DayOffsetMiddleware() gin.HandlerFunc {
	spec := QueryIntVar{
		Key:        "offset",
		Default:    0,
		ErrInvalid: "offset must be an integer",
	}
	return func(c *gin.Context) {
		if raw := c.Query("date"); raw != "" {
			day, err := time.ParseInLocation("2006-01-02", raw, time.Local)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
				c.Abort()
				return
			}
			c.Set(ginContextKeyDayOffset, OffsetForDate(day, clock.Now()))
			c.Next()
			return
		}
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

// OffsetForDate is the day offset (0 = today, 1 = yesterday) of calendar day `day` relative to
// `now`, both read in their own location's calendar. Counted in calendar days, so DST doesn't skew it.
func OffsetForDate(day, now time.Time) int {
	d := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	n := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return int(n.Sub(d).Hours() / 24)
}
