package utils

import (
	"strings"
	"time"

	"zion-english/internal/constants"
)

func DatePHT(t time.Time) string {
	return t.In(constants.LocationPHT).Format(constants.DateLayout)
}

func TodayPHT() string {
	return DatePHT(time.Now())
}

func ParseDatePHT(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation(constants.DateLayout, value, constants.LocationPHT)
	if err != nil {
		return nil, err
	}
	utc := t.UTC()
	return &utc, nil
}

// NormalizeDatePHT trims, parses a PHT calendar date, and returns it formatted as YYYY-MM-DD.
// On parse failure the original trimmed input is returned.
func NormalizeDatePHT(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	t, err := ParseDatePHT(value)
	if err != nil || t == nil {
		return value
	}
	return DatePHT(*t)
}
