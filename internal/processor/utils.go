package processor

import (
	"errors"
	"strings"
	"time"
)

// ColumnLetterToIndex converts a column letter (e.g., "A", "B") to a 0-based index
func ColumnLetterToIndex(col string) int {
	col = strings.ToUpper(strings.TrimSpace(col))
	if len(col) == 0 {
		return -1
	}
	return int(col[0] - 'A')
}

func ParseDateString(dateStr string) (*time.Time, error) {
	return ParseDateStringWithYear(dateStr, 0)
}

func ParseDateStringWithYear(dateStr string, year int) (*time.Time, error) {
	dateStr = strings.Trim(dateStr, `"`)
	dateStr = strings.TrimSpace(dateStr)

	normalized := normalizeDateString(dateStr)

	parsedDate, err := time.Parse("January 2, 2006", normalized)
	if err != nil {
		parsedDate, err = time.Parse("January 02, 2006", normalized)
	}
	if err != nil {
		// Try "Mon-DD" format (e.g., "Oct-16", "Oct.-16")
		parsedDate, err = time.Parse("Jan-02", normalized)
		if err != nil {
			parsedDate, err = time.Parse("Jan-2", normalized)
		}
		if err != nil {
			// Try "DD-Mon" format (e.g., "16-Jan", "16-Jan.")
			parsedDate, err = time.Parse("02-Jan", normalized)
			if err != nil {
				parsedDate, err = time.Parse("2-Jan", normalized)
			}
		}
		if err == nil {
			useYear := year
			if useYear == 0 {
				useYear = max(time.Now().Year(), 2025)
			}
			parsedDate = time.Date(useYear, parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
		}
	}
	if err != nil {
		return nil, err
	}
	return &parsedDate, nil
}

func ParseDateFromRecord(record []string, year int) (*time.Time, error) {
	if len(record) == 0 {
		return nil, errors.New("empty record")
	}
	return ParseDateStringWithYear(record[0], year)
}

func normalizeDateString(dateStr string) string {
	result := dateStr

	dayNames := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, day := range dayNames {
		result = strings.TrimPrefix(result, day+", ")
		result = strings.TrimPrefix(result, day+",")
		result = strings.TrimPrefix(result, day+" , ")
		result = strings.TrimPrefix(result, day+" ,")
	}

	result = strings.ReplaceAll(result, "Jan.", "Jan")
	result = strings.ReplaceAll(result, "Feb.", "Feb")
	result = strings.ReplaceAll(result, "Mar.", "Mar")
	result = strings.ReplaceAll(result, "Apr.", "Apr")
	result = strings.ReplaceAll(result, "May.", "May")
	result = strings.ReplaceAll(result, "Jun.", "Jun")
	result = strings.ReplaceAll(result, "Jul.", "Jul")
	result = strings.ReplaceAll(result, "Aug.", "Aug")
	result = strings.ReplaceAll(result, "Sep.", "Sep")
	result = strings.ReplaceAll(result, "Sept.", "Sep")
	result = strings.ReplaceAll(result, "Oct.", "Oct")
	result = strings.ReplaceAll(result, "Nov.", "Nov")
	result = strings.ReplaceAll(result, "Dec.", "Dec")

	result = strings.Join(strings.Fields(result), " ")
	return strings.TrimSpace(result)
}

// ParseTimeString parses a time string and combines it with a date
func ParseTimeString(timeStr string, date time.Time) (time.Time, error) {
	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		parsedTime, err = time.Parse("15:04:05", timeStr)
		if err != nil {
			return time.Time{}, err
		}
	}

	result := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		parsedTime.Second(),
		0,
		date.Location(),
	)

	return result, nil
}

func ParseClassTime(record string, date time.Time) (*time.Time, error) {
	timeStr := strings.TrimSpace(record)
	if timeStr != "" {
		parsedTime, err := ParseTimeString(timeStr, date)
		if err != nil {
			return nil, err
		}
		return &parsedTime, nil
	}
	return nil, errors.New("invalid time")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
