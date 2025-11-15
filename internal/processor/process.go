package processor

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"zion-english/internal/logs"

	"go.uber.org/zap"
)

func ProcessCSVFile(filePath string, startDate, endDate time.Time, colIndices ColumnIndices) ([]ClassRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var endIdx int = -1
	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		if len(record) <= 0 {
			continue
		}
		parsedDate, err := ParseDateFromRecord(record, endDate.Year())
		if err != nil {
			continue
		}

		if parsedDate.Year() == endDate.Year() &&
			parsedDate.Month() == endDate.Month() &&
			parsedDate.Day() == endDate.Day() {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return nil, fmt.Errorf("end date not found in CSV")
	}

	var startIdx int = -1
	for i := endIdx; i >= 0; i-- {
		record := records[i]
		if len(record) <= 0 {
			continue
		}
		parsedDate, err := ParseDateFromRecord(record, endDate.Year())
		if err != nil {
			continue
		}

		if parsedDate.Month() == startDate.Month() &&
			parsedDate.Day() == startDate.Day() &&
			parsedDate.Year() == endDate.Year() {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return nil, fmt.Errorf("start date not found in CSV")
	}

	logs.Log().Info(
		"Date row indices",
		zap.Int("startIndex", startIdx),
		zap.Int("endIndex", endIdx),
	)

	var teacherRecords []ClassRecord
	type dateRow struct {
		index int
		date  time.Time
	}
	var dateRows []dateRow

	for i := startIdx; i <= endIdx; i++ {
		record := records[i]
		if len(record) == 0 || strings.TrimSpace(record[0]) == "" {
			continue
		}
		parsedDate, err := ParseDateFromRecord(record, endDate.Year())
		if err == nil {
			if parsedDate.After(endDate) {
				break
			}
			if !parsedDate.Before(startDate) {
				dateRows = append(dateRows, dateRow{index: i, date: *parsedDate})
			}
		}
	}

	for i := startIdx + 1; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 || strings.TrimSpace(record[0]) == "" {
			continue
		}

		parsedDate, err := ParseDateFromRecord(record, endDate.Year())
		if err == nil {
			if parsedDate.After(endDate) {
				break
			}
			continue
		}

		var currentDate time.Time
		found := false
		for j := len(dateRows) - 1; j >= 0; j-- {
			if dateRows[j].index < i {
				currentDate = dateRows[j].date
				found = true
				break
			}
		}
		if !found && len(dateRows) > 0 {
			lastDateRow := dateRows[len(dateRows)-1]
			if lastDateRow.index == endIdx && i > endIdx {
				currentDate = lastDateRow.date
				found = true
			}
		}
		if !found {
			continue
		}

		if currentDate.Before(startDate) || currentDate.After(endDate) {
			continue
		}

		teacherRec := parseTeacherRecord(record, currentDate, colIndices)
		if teacherRec != nil {
			teacherRec.OriginalRowIndex = i + 1
			teacherRecords = append(teacherRecords, *teacherRec)
		}
	}

	sortedTeacherRecords := make([]ClassRecord, len(teacherRecords))
	copy(sortedTeacherRecords, teacherRecords)
	sort.Slice(sortedTeacherRecords, func(i, j int) bool {
		if sortedTeacherRecords[i].Name != sortedTeacherRecords[j].Name {
			return sortedTeacherRecords[i].Name < sortedTeacherRecords[j].Name
		}
		dateI, errI := ParseDateString(sortedTeacherRecords[i].Date)
		dateJ, errJ := ParseDateString(sortedTeacherRecords[j].Date)
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		return dateI.Before(*dateJ)
	})

	return sortedTeacherRecords, nil
}

func parseTeacherRecord(record []string, date time.Time, colIndices ColumnIndices) *ClassRecord {
	if len(record) == 0 {
		return nil
	}

	// Name column
	name := ""
	if colIndices.Name >= 0 && len(record) > colIndices.Name {
		name = strings.TrimSpace(record[colIndices.Name])
	}
	if name == "" {
		return nil // Skip records without a name
	}

	// Duration column - Time in minutes (e.g., "25 min.", "25 MINS")
	timeMin := 0
	if colIndices.Duration >= 0 && len(record) > colIndices.Duration {
		timeStr := strings.TrimSpace(record[colIndices.Duration])
		timeStr = strings.ToUpper(timeStr)
		timeStr = strings.TrimSuffix(timeStr, ".")
		timeStr = strings.TrimSuffix(timeStr, "MINS")
		timeStr = strings.TrimSuffix(timeStr, "MIN")
		timeStr = strings.TrimSpace(timeStr)
		if parsed, err := strconv.Atoi(timeStr); err == nil {
			timeMin = parsed
		}
	}

	// Rate column
	rate := 0.0
	if colIndices.Rate >= 0 && len(record) > colIndices.Rate {
		rateStr := strings.TrimSpace(record[colIndices.Rate])
		if rateStr != "-" && rateStr != "" {
			if parsed, err := strconv.ParseFloat(rateStr, 64); err == nil {
				rate = parsed
			}
		}
	}

	// Start time column
	var startTime *time.Time
	if colIndices.StartTime >= 0 && len(record) > colIndices.StartTime {
		parsedTime, err := ParseClassTime(record[colIndices.StartTime], date)
		if err == nil {
			startTime = parsedTime
		}
	}

	// End time column
	var endTime *time.Time
	if colIndices.EndTime >= 0 && len(record) > colIndices.EndTime {
		parsedTime, err := ParseClassTime(record[colIndices.EndTime], date)
		if err == nil {
			endTime = parsedTime
		}
	}

	// Google link column
	googleLink := ""
	if colIndices.Link >= 0 && len(record) > colIndices.Link {
		googleLink = strings.TrimSpace(record[colIndices.Link])
	}

	// Status column
	status := ""
	if colIndices.Status >= 0 && len(record) > colIndices.Status {
		status = strings.TrimSpace(record[colIndices.Status])
	}

	return &ClassRecord{
		Name:              name,
		DurationInMinutes: timeMin,
		Rate:              rate,
		StartTime:         startTime,
		EndTime:           endTime,
		GoogleLink:        googleLink,
		Status:            status,
		Date:              date.Format("January 2, 2006"),
	}
}
