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
	return CutoffDatesFromPreset(active)
}

// InCutoffRepeatWindow reports whether today (YYYY-MM-DD, PHT) falls in the repeat window
// of N days on or before the active payroll cutoff end (inclusive). N=3 with cutoff end Sep 15
// includes Sep 13, 14, and 15.
func InCutoffRepeatWindow(today string, n int) bool {
	_, endDate := ActiveCutoffDates()
	return CutoffRepeatWindowContains(today, endDate, n)
}

// CutoffRepeatWindowContains reports whether today is within N days on or before cutoffEnd (inclusive).
func CutoffRepeatWindowContains(today, cutoffEnd string, n int) bool {
	if n < 0 || n > 10 {
		return false
	}
	cutoffEnd = strings.TrimSpace(cutoffEnd)
	if cutoffEnd == "" {
		return false
	}
	end, err := time.ParseInLocation(constants.DateLayout, cutoffEnd, constants.LocationPHT)
	if err != nil {
		return false
	}
	offset := 0
	if n > 0 {
		offset = n - 1
	}
	start := end.AddDate(0, 0, -offset)
	startDay := start.Format(constants.DateLayout)
	return today >= startDay && today <= cutoffEnd
}

// NextCutoffDates returns the payroll cutoff that follows the given period.
// Returns ok false when the dates do not match a standard first or second cutoff.
func NextCutoffDates(startDate, endDate string) (nextStart, nextEnd string, ok bool) {
	startT, err := time.ParseInLocation(constants.DateLayout, startDate, constants.LocationPHT)
	if err != nil {
		return "", "", false
	}
	if _, err := time.ParseInLocation(constants.DateLayout, endDate, constants.LocationPHT); err != nil {
		return "", "", false
	}

	firstPreset, secondPreset := CutoffRangeForMonth(startT.Year(), startT.Month())
	firstStart, firstEnd := CutoffDatesFromPreset(firstPreset)
	secondStart, secondEnd := CutoffDatesFromPreset(secondPreset)

	if startDate == firstStart && endDate == firstEnd {
		return secondStart, secondEnd, true
	}
	if startDate == secondStart && endDate == secondEnd {
		nextMonth := startT.AddDate(0, 1, 0)
		nextFirst, _ := CutoffRangeForMonth(nextMonth.Year(), nextMonth.Month())
		nextStart, nextEnd = CutoffDatesFromPreset(nextFirst)
		return nextStart, nextEnd, nextStart != "" && nextEnd != ""
	}
	return "", "", false
}

// CutoffDatesFromPreset parses a cutoff preset (YYYY-MM-DD|YYYY-MM-DD) into start and end dates.
func CutoffDatesFromPreset(preset string) (startDate, endDate string) {
	parts := splitCutoff(preset)
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

// FinalReportRangeFromResignationDate returns the payroll cutoff start through the resignation date (PHT).
func FinalReportRangeFromResignationDate(resignDate string) (startDate, endDate string, err error) {
	resignDate = strings.TrimSpace(resignDate)
	if resignDate == "" {
		return "", "", ErrResignationDateOutsideCutoff
	}
	t, err := time.ParseInLocation(constants.DateLayout, resignDate, constants.LocationPHT)
	if err != nil {
		return "", "", err
	}
	day := t.Format(constants.DateLayout)
	firstPreset, secondPreset := CutoffRangeForMonth(t.Year(), t.Month())
	firstStart, firstEnd := CutoffDatesFromPreset(firstPreset)
	secondStart, secondEnd := CutoffDatesFromPreset(secondPreset)
	if day >= firstStart && day <= firstEnd {
		return firstStart, day, nil
	}
	if day >= secondStart && day <= secondEnd {
		return secondStart, day, nil
	}
	return "", "", ErrResignationDateOutsideCutoff
}
