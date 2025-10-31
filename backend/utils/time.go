package utils

import "time"

func ZerodTime(offset int) time.Time {
    now := time.Now() // local time
    return time.Date(now.Year(), now.Month(), now.Day() - offset, 0, 0, 0, 0, now.Location())
}