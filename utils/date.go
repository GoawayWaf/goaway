package utils

import (
	"time"
	_"fmt"
)

const MYSQLDATETIME = "2006-01-02 15:04:05"
const LOGDATEFORMAT = "02/Jan/2006:15:04:05 +0000"

func GetTimeSinceSeconds(d time.Duration) time.Time {
	now := time.Now()
	// time.Add(-(time.Seconds * d))
	timeSinceSeconds := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second()-int(d.Seconds()),
		now.Nanosecond(),
		now.Location(),
	)
	return timeSinceSeconds
}
