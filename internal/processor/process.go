package processor

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
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

	var startIdx int = -1
	for i, record := range records {
		if len(record) <= 0 {
			continue
		}
		parsedDate, err := ParseDateFromRecord(record, startDate.Year())
		if err != nil {
			continue
		}

		if parsedDate.Year() == startDate.Year() &&
			parsedDate.Month() == startDate.Month() &&
			parsedDate.Day() == startDate.Day() {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return nil, fmt.Errorf("start date not found in CSV")
	}

	var endIdx int = -1
	for i, record := range records {
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

	var teacherRecords []ClassRecord
	currentDate := startDate
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
			currentDate = *parsedDate
			continue
		}

		if currentDate.After(endDate) {
			break
		}

		teacherRec := parseTeacherRecord(record, currentDate, colIndices)
		if teacherRec != nil {
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
		// Normalize to uppercase for consistent parsing
		timeStr = strings.ToUpper(timeStr)
		// Remove common suffixes: "MINS", "MIN", "MINS.", "MIN.", etc.
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

func SaveRecordsToCSV(records []ClassRecord, outputPath string, colIndices ColumnIndices) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	header := []string{"Name", "Date", "Duration", "Rate"}
	if colIndices.StartTime >= 0 {
		header = append(header, "StartTime")
	}
	if colIndices.EndTime >= 0 {
		header = append(header, "EndTime")
	}
	if colIndices.Link >= 0 {
		header = append(header, "GoogleLink")
	}
	header = append(header, "Status")

	if err := writer.Write(header); err != nil {
		return err
	}

	var total float64
	for _, rec := range records {
		record := []string{
			rec.Name,
			rec.Date,
			strconv.Itoa(rec.DurationInMinutes),
			strconv.FormatFloat(rec.Rate, 'f', 2, 64),
		}

		if colIndices.StartTime >= 0 {
			startTimeStr := ""
			if rec.StartTime != nil {
				startTimeStr = rec.StartTime.Format("15:04")
			}
			record = append(record, startTimeStr)
		}

		if colIndices.EndTime >= 0 {
			endTimeStr := ""
			if rec.EndTime != nil {
				endTimeStr = rec.EndTime.Format("15:04")
			}
			record = append(record, endTimeStr)
		}

		if colIndices.Link >= 0 {
			record = append(record, rec.GoogleLink)
		}

		record = append(record, rec.Status)
		total += rec.Rate

		if err := writer.Write(record); err != nil {
			return err
		}
	}
	if err := writer.Write([]string{fmt.Sprintf("TOTAL: %.2f", total)}); err != nil {
		return err
	}

	return nil
}
