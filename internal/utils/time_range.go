package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"zion-english/internal/constants"
)

var (
	ErrTimeRequired      = errors.New("time is required")
	ErrInvalidTimeFormat = errors.New("invalid time format")
	ErrInvalidStartTime  = errors.New("invalid start time")
	ErrInvalidEndTime    = errors.New("invalid end time")
	ErrEndBeforeStart    = errors.New("end time must be after start time")
)

func ParseTimeHM(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, ErrTimeRequired
	}
	for _, layout := range []string{constants.TimeHMSecondsLayout, constants.TimeHMLayout} {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidTimeFormat, value)
}

func DurationMinutesFromRange(start, end string) (int64, error) {
	startT, err := ParseTimeHM(start)
	if err != nil {
		return 0, ErrInvalidStartTime
	}
	endT, err := ParseTimeHM(end)
	if err != nil {
		return 0, ErrInvalidEndTime
	}

	startMins := startT.Hour()*60 + startT.Minute()
	endMins := endT.Hour()*60 + endT.Minute()
	if endMins <= startMins {
		return 0, ErrEndBeforeStart
	}
	return int64(endMins - startMins), nil
}

func EndTimeFromStartAndDuration(start string, durationMinutes int64) string {
	startT, err := ParseTimeHM(start)
	if err != nil || durationMinutes <= 0 {
		return ""
	}
	total := startT.Hour()*60 + startT.Minute() + int(durationMinutes)
	h := (total / 60) % 24
	m := total % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func FormatDurationMinutes(minutes int64) string {
	if minutes < 60 {
		return fmt.Sprintf("%d min", minutes)
	}
	h := minutes / 60
	m := minutes % 60
	if m == 0 {
		return fmt.Sprintf("%d hr", h)
	}
	return fmt.Sprintf("%d hr %d min", h, m)
}
