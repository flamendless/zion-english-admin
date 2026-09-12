package utils

import (
	"database/sql"
	"sort"
	"strings"
	"time"

	"zion-english/internal/constants"
)

func DatePHT(t time.Time) string {
	return t.In(constants.LocationPHT).Format(constants.DateLayout)
}

func DateTimePHT(t time.Time) string {
	return t.In(constants.LocationPHT).Format(constants.DateTimeLayout)
}

func DateTimeSecondsPHT(t time.Time) string {
	return t.In(constants.LocationPHT).Format(constants.DateTimeSecondsLayout)
}

func FormatNullDateTimePHT(t sql.NullTime) string {
	if !t.Valid {
		return "-"
	}
	return DateTimePHT(t.Time)
}

func FormatNullDateTimeSecondsPHT(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return DateTimeSecondsPHT(t.Time)
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

func IsDateTimeInPastPHT(date, startTime string) bool {
	date = strings.TrimSpace(date)
	startTime = strings.TrimSpace(startTime)
	if date == "" || startTime == "" {
		return false
	}
	day, err := time.ParseInLocation(constants.DateLayout, date, constants.LocationPHT)
	if err != nil {
		return false
	}
	hm, err := ParseTimeHM(startTime)
	if err != nil {
		return false
	}
	dt := time.Date(day.Year(), day.Month(), day.Day(), hm.Hour(), hm.Minute(), 0, 0, constants.LocationPHT)
	return dt.Before(time.Now().In(constants.LocationPHT))
}

type compactDateRun struct {
	start time.Time
	end   time.Time
}

func FormatCompactDateRanges(dates []string) string {
	if len(dates) == 0 {
		return "-"
	}

	seen := make(map[string]struct{}, len(dates))
	parsed := make([]time.Time, 0, len(dates))
	for _, raw := range dates {
		date := strings.TrimSpace(raw)
		if date == "" {
			continue
		}
		if _, ok := seen[date]; ok {
			continue
		}
		t, err := time.ParseInLocation(constants.DateLayout, date, constants.LocationPHT)
		if err != nil {
			continue
		}
		seen[date] = struct{}{}
		parsed = append(parsed, t)
	}
	if len(parsed) == 0 {
		return "-"
	}

	sort.Slice(parsed, func(i, j int) bool {
		return parsed[i].Before(parsed[j])
	})

	runs := []compactDateRun{{start: parsed[0], end: parsed[0]}}
	for i := 1; i < len(parsed); i++ {
		prev := parsed[i-1]
		cur := parsed[i]
		if DatePHT(prev.AddDate(0, 0, 1)) == DatePHT(cur) {
			runs[len(runs)-1].end = cur
			continue
		}
		runs = append(runs, compactDateRun{start: cur, end: cur})
	}

	allSameYear := true
	year := parsed[0].Year()
	for _, t := range parsed {
		if t.Year() != year {
			allSameYear = false
			break
		}
	}

	segments := formatCompactDateSegments(runs, allSameYear)
	if len(segments) == 0 {
		return "-"
	}
	out := strings.Join(segments, ", ")
	if allSameYear {
		return out + ", " + parsed[0].Format("2006")
	}
	return out
}

func formatCompactDateSegments(runs []compactDateRun, allSameYear bool) []string {
	segments := make([]string, 0, len(runs))
	var prevMonth time.Month
	var prevYear int
	hasPrev := false

	for _, run := range runs {
		segment := formatCompactDateSegment(run.start, run.end, allSameYear, hasPrev, prevYear, prevMonth)
		segments = append(segments, segment)
		prevMonth = run.end.Month()
		prevYear = run.end.Year()
		hasPrev = true
	}
	return segments
}

func formatCompactDateSegment(start, end time.Time, allSameYear, hasPrevMonth bool, prevYear int, prevMonth time.Month) string {
	sameMonthAsPrev := hasPrevMonth && allSameYear && start.Year() == prevYear && start.Month() == prevMonth

	if start.Equal(end) {
		if sameMonthAsPrev {
			return start.Format("2")
		}
		if allSameYear {
			return start.Format("Jan 2")
		}
		return start.Format("Jan 2, 2006")
	}

	if allSameYear && start.Year() == end.Year() && start.Month() == end.Month() {
		if sameMonthAsPrev {
			return start.Format("2") + " - " + end.Format("2")
		}
		return start.Format("Jan 2") + " - " + end.Format("2")
	}
	if allSameYear && start.Year() == end.Year() {
		return start.Format("Jan 2") + " - " + end.Format("Jan 2")
	}
	return start.Format("Jan 2, 2006") + " - " + end.Format("Jan 2, 2006")
}
