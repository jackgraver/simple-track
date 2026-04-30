package common

import (
	"time"

	"be-simpletracker/internal/utils"
)

// ParseDateString parses YYYY-MM-DD in local time, or returns today at midnight if empty.
func ParseDateString(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return utils.ZerodTime(0), nil
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
}
