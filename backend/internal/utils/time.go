package utils

import (
	"be-simpletracker/internal/clock"
	"time"
)

func ZerodTime(offset int) time.Time {
	now := clock.Now() // local time
	return time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, now.Location())
}
