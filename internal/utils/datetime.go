package utils

import (
	"strings"
	"time"

	"zion-english/internal/constants"
)

func DatePHT(t time.Time) string {
	return t.In(constants.LocationPHT).Format(constants.DateLayout)
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
