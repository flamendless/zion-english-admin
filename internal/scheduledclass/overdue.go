package scheduledclass

import (
	"time"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

func ScheduledEndAtPHT(scheduledDate, startTime string, durationMinutes int64) (time.Time, bool) {
	if scheduledDate == "" {
		return time.Time{}, false
	}
	date, err := time.ParseInLocation(constants.DateLayout, scheduledDate, constants.LocationPHT)
	if err != nil {
		return time.Time{}, false
	}
	if startTime != "" && durationMinutes > 0 {
		startMins, err := utils.MinutesSinceMidnight(startTime)
		if err != nil {
			return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, constants.LocationPHT), true
		}
		total := int(startMins) + int(durationMinutes)
		return time.Date(date.Year(), date.Month(), date.Day(), total/60%24, total%60, 0, 0, constants.LocationPHT), true
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, constants.LocationPHT), true
}

func IsOverdue(status constants.ScheduledClassStatus, scheduledDate, startTime string, durationMinutes int64, now time.Time, gracePeriodMinutes int64) bool {
	if status != constants.ScheduledClassStatusScheduled {
		return false
	}
	endAt, ok := ScheduledEndAtPHT(scheduledDate, startTime, durationMinutes)
	if !ok {
		return false
	}
	if gracePeriodMinutes < 0 {
		gracePeriodMinutes = 0
	}
	threshold := endAt.Add(time.Duration(gracePeriodMinutes) * time.Minute)
	return now.In(constants.LocationPHT).After(threshold)
}
