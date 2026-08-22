package utils

import (
	"time"
	"zion-english/internal/constants"
)

// CurrentCutoffRange returns the first cutoff, second cutoff, and active cutoff date ranges
// for the current calendar month in Philippines time (YYYY-MM-DD|YYYY-MM-DD format).
func CurrentCutoffRange() (firstCutoff, secondCutoff, activeCutoff string) {
	now := time.Now().In(constants.LocationPHT)
	year, month, day := now.Date()
	lastOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, constants.LocationPHT).AddDate(0, 1, -1).Day()

	a := time.Date(year, month, 1, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)
	b := time.Date(year, month, 15, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)
	c := time.Date(year, month, 16, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)
	d := time.Date(year, month, lastOfMonth, 0, 0, 0, 0, constants.LocationPHT).Format(constants.DateLayout)

	firstCutoff = a + "|" + b
	secondCutoff = c + "|" + d
	if day <= 15 {
		activeCutoff = firstCutoff
	} else {
		activeCutoff = secondCutoff
	}
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
