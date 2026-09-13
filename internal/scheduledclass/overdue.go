package scheduledclass

import (
	"time"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

func scheduledDayEndPHT(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, constants.LocationPHT)
}

func ScheduledEndAtPHT(scheduledDate, startTime string, durationMinutes int64) (time.Time, bool) {
	if scheduledDate == "" {
		return time.Time{}, false
	}
	date, err := time.ParseInLocation(constants.DateLayout, scheduledDate, constants.LocationPHT)
	if err != nil {
		return time.Time{}, false
	}
	if startTime != "" && durationMinutes > 0 {
		startAt, err := utils.ParseTimeHM(startTime)
		if err != nil {
			return scheduledDayEndPHT(date), true
		}
		classStart := time.Date(
			date.Year(), date.Month(), date.Day(),
			startAt.Hour(), startAt.Minute(), 0, 0,
			constants.LocationPHT,
		)
		return classStart.Add(time.Duration(durationMinutes) * time.Minute), true
	}
	return scheduledDayEndPHT(date), true
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
