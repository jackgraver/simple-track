package common

import (
	"be-simpletracker/internal/clock"
	"time"
)

// ParseDateString parses YYYY-MM-DD in local time, or returns today at midnight if empty.
func ParseDateString(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return clock.ZerodTime(0), nil
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
}
