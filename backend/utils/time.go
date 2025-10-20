package utils

import "time"

func ZerodTime() time.Time {
    now := time.Now() // local time
    return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}