package processor

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func ColumnLetterToIndex(col string) int {
	col = strings.ToUpper(strings.TrimSpace(col))
	if len(col) == 0 {
		return -1
	}
	return int(col[0] - 'A')
}

var ErrInvalidSheetTemplate = errors.New("template must be four comma-separated Excel column letters (e.g. A,B,C,G)")

func ValidateSheetTemplate(template string) error {
	template = strings.TrimSpace(template)
	if template == "" {
		return nil
	}
	parts := strings.Split(template, ",")
	if len(parts) != 4 {
		return ErrInvalidSheetTemplate
	}
	for _, part := range parts {
		part = strings.ToUpper(strings.TrimSpace(part))
		if len(part) != 1 || part[0] < 'A' || part[0] > 'Z' {
			return ErrInvalidSheetTemplate
		}
	}
	return nil
}

func ParseDateString(dateStr string) (*time.Time, error) {
	return ParseDateStringWithYear(dateStr, 0)
}

func ParseDateStringWithYear(dateStr string, year int) (*time.Time, error) {
	dateStr = strings.Trim(dateStr, `"`)
	dateStr = strings.TrimSpace(dateStr)

	// Try parsing "MM/DD/YY" or "M/D/YY" format (e.g., "11/30/25", "5/16/25")
	if strings.Contains(dateStr, "/") {
		parts := strings.Split(dateStr, "/")
		if len(parts) == 3 {
			month, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			day, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			yearStr := strings.TrimSpace(parts[2])

			if err1 == nil && err2 == nil && month >= 1 && month <= 12 && day >= 1 && day <= 31 {
				// Parse year from string
				yearVal, err3 := strconv.Atoi(yearStr)
				if err3 == nil {
					useYear := yearVal
					if len(yearStr) == 2 {
						// Convert 2-digit year to 4-digit (25 -> 2025, assuming 2000s)
						useYear = 2000 + yearVal
					} else if len(yearStr) == 4 {
						useYear = yearVal
					} else {
						// Fallback to parameter year or current year
						if year != 0 {
							useYear = year
						} else {
							useYear = max(time.Now().Year(), 2025)
						}
					}

					parsedDate := time.Date(useYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
					return &parsedDate, nil
				}
			}
		}
	}

	normalized := normalizeDateString(dateStr)

	if parsedDate, err := time.Parse("2006-01-02", normalized); err == nil {
		return &parsedDate, err
	}

	parsedDate, err := time.Parse("January 2, 2006", normalized)
	if err != nil {
		parsedDate, err = time.Parse("January 02, 2006", normalized)
	}
	if err != nil {
		// Try "January 2" format without comma (e.g., "November 29", "November 2")
		parsedDate, err = time.Parse("January 2", normalized)
		if err != nil {
			parsedDate, err = time.Parse("January 02", normalized)
		}
	}
	if err != nil {
		// Try "Mon DD" format (e.g., "Nov 29", "Nov 2")
		parsedDate, err = time.Parse("Jan 2", normalized)
		if err != nil {
			parsedDate, err = time.Parse("Jan 02", normalized)
		}
	}
	if err != nil {
		// Try "Mon-DD" format (e.g., "Oct-16", "Oct.-16")
		parsedDate, err = time.Parse("Jan-02", normalized)
		if err != nil {
			parsedDate, err = time.Parse("Jan-2", normalized)
		}
	}
	if err != nil {
		// Try "DD-Mon" format (e.g., "16-Jan", "16-Jan.")
		parsedDate, err = time.Parse("02-Jan", normalized)
		if err != nil {
			parsedDate, err = time.Parse("2-Jan", normalized)
		}
	}

	// If parsing succeeded but no year was in the format, use the current year
	// Formats without year (like "Jan 2", "January 2", "Jan-2", "2-Jan") will have year <= 1
	if err == nil && parsedDate.Year() <= 1 {
		useYear := time.Now().Year()
		parsedDate = time.Date(useYear, parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
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
