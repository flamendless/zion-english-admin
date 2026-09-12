package utils

import (
	"strings"
	"time"
	"zion-english/internal/constants"
)

// CutoffRangeForMonth returns first and second payroll cutoff presets for a calendar month
// in Philippines time (YYYY-MM-DD|YYYY-MM-DD format).
func CutoffRangeForMonth(year int, month time.Month) (firstCutoff, secondCutoff string) {
	lastOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, constants.LocationPHT).AddDate(0, 1, -1).Day()

	a := time.Date(year, month, 1, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)
	b := time.Date(year, month, 15, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)
	c := time.Date(year, month, 16, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)
	d := time.Date(year, month, lastOfMonth, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)

	return a + "|" + b, c + "|" + d
}

// ActiveCutoffForMonth returns the payroll cutoff preset for a month. For the current month
// it matches today's cutoff; for other months it defaults to the first cutoff.
func ActiveCutoffForMonth(year int, month time.Month) string {
	first, second := CutoffRangeForMonth(year, month)
	now := time.Now().In(constants.LocationPHT)
	if year != now.Year() || month != now.Month() {
		return first
	}
	if now.Day() <= 15 {
		return first
	}
	return second
}

// CurrentMonthPHT returns the current calendar month in Philippines time (YYYY-MM).
func CurrentMonthPHT() string {
	return time.Now().In(constants.LocationPHT).Format(constants.MonthLayout)
}

// ParseMonthPHT parses a YYYY-MM value as a Philippines calendar month.
func ParseMonthPHT(value string) (year int, month time.Month, ok bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, 0, false
	}
	t, err := time.ParseInLocation(constants.MonthLayout, value, constants.LocationPHT)
	if err != nil {
		return 0, 0, false
	}
	return t.Year(), t.Month(), true
}

// CurrentCutoffRange returns the first cutoff, second cutoff, and active cutoff date ranges
// for the current calendar month in Philippines time (YYYY-MM-DD|YYYY-MM-DD format).
func CurrentCutoffRange() (firstCutoff, secondCutoff, activeCutoff string) {
	now := time.Now().In(constants.LocationPHT)
	firstCutoff, secondCutoff = CutoffRangeForMonth(now.Year(), now.Month())
	activeCutoff = ActiveCutoffForMonth(now.Year(), now.Month())
	return firstCutoff, secondCutoff, activeCutoff
}

// ActiveCutoffDates returns the start and end date strings for the current payroll cutoff.
func ActiveCutoffDates() (startDate, endDate string) {
	_, _, active := CurrentCutoffRange()
	parts := splitCutoff(active)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func splitCutoff(value string) []string {
	for i := 0; i < len(value); i++ {
		if value[i] == '|' {
			return []string{value[:i], value[i+1:]}
		}
	}
	return nil
}
