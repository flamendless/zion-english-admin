package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"zion-english/internal/logs"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type ParseTeacherFlags struct {
	filepath     string
	output       string
	startDate    string
	endDate      string
	nameCol      string
	durationCol  string
	rateCol      string
	statusCol    string
	startTimeCol string
	endTimeCol   string
	linkCol      string
}

var parseTeacherFlags ParseTeacherFlags

type ClassRecord struct {
	Name              string
	DurationInMinutes int
	Rate              float64
	StartTime         *time.Time
	EndTime           *time.Time
	GoogleLink        string
	Status            string
	Date              string
}

func init() {
	f := cmdParseTeacher.Flags
	f().StringVarP(&parseTeacherFlags.filepath, "filepath", "f", "", "Filepath to process (csv)")
	f().StringVarP(&parseTeacherFlags.output, "output", "o", "", "Output path")
	f().StringVarP(&parseTeacherFlags.startDate, "startDate", "s", "", "Start date in file to process (format: 'January 2, 2006' or 'January 02, 2006')")
	f().StringVarP(&parseTeacherFlags.endDate, "endDate", "e", "", "End date in file to process (format: 'January 2, 2006' or 'January 02, 2006')")
	f().StringVarP(&parseTeacherFlags.nameCol, "nameCol", "n", "A", "Column for name (e.g., A, B, C)")
	f().StringVarP(&parseTeacherFlags.durationCol, "durationCol", "d", "B", "Column for duration/time (e.g., A, B, C)")
	f().StringVarP(&parseTeacherFlags.rateCol, "rateCol", "r", "C", "Column for rate (e.g., A, B, C)")
	f().StringVarP(&parseTeacherFlags.statusCol, "statusCol", "t", "G", "Column for status (e.g., A, B, C)")
	f().StringVarP(&parseTeacherFlags.startTimeCol, "startTimeCol", "", "", "Column for start time (e.g., A, B, C) - optional")
	f().StringVarP(&parseTeacherFlags.endTimeCol, "endTimeCol", "", "", "Column for end time (e.g., A, B, C) - optional")
	f().StringVarP(&parseTeacherFlags.linkCol, "linkCol", "", "", "Column for Google link (e.g., A, B, C) - optional")
	if err := cmdParseTeacher.MarkFlagRequired("filepath"); err != nil {
		panic(err)
	}
	if err := cmdParseTeacher.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
	if err := cmdParseTeacher.MarkFlagRequired("startDate"); err != nil {
		panic(err)
	}
	if err := cmdParseTeacher.MarkFlagRequired("endDate"); err != nil {
		panic(err)
	}
	if err := cmdParseTeacher.MarkFlagRequired("nameCol"); err != nil {
		panic(err)
	}
	if err := cmdParseTeacher.MarkFlagRequired("durationCol"); err != nil {
		panic(err)
	}
	if err := cmdParseTeacher.MarkFlagRequired("rateCol"); err != nil {
		panic(err)
	}
	if err := cmdParseTeacher.MarkFlagRequired("statusCol"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(cmdParseTeacher)
}

var cmdParseTeacher = &cobra.Command{
	Use:   "parse_teacher",
	Short: "Parse teacher",
	Run: func(cmd *cobra.Command, args []string) {
		startDateStr := strings.TrimSpace(parseTeacherFlags.startDate)
		parsedStartDate, err := parseDateString(startDateStr)
		if err != nil {
			logs.Log().Error("Failed to parse start date", zap.String("date", startDateStr), zap.Error(err))
			return
		}
		targetStartDate := *parsedStartDate

		endDateStr := strings.TrimSpace(parseTeacherFlags.endDate)
		parsedEndDate, err := parseDateString(endDateStr)
		if err != nil {
			logs.Log().Error("Failed to parse end date", zap.String("date", endDateStr), zap.Error(err))
			return
		}
		targetEndDate := *parsedEndDate

		logs.Log().Info(
			"Processing teacher",
			zap.String("filepath", parseTeacherFlags.filepath),
			zap.String("output", parseTeacherFlags.output),
			zap.Time("start date", targetStartDate),
			zap.Time("end date", targetEndDate),
		)

		file, err := os.Open(parseTeacherFlags.filepath)
		if err != nil {
			logs.Log().Error("Failed to open file", zap.Error(err))
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			logs.Log().Error("Failed to read CSV", zap.Error(err))
			return
		}

		var startIdx int = -1
		for i, record := range records {
			if len(record) <= 0 {
				continue
			}
			parsedDate, err := parseDate(record, targetStartDate.Year())
			if err != nil {
				continue
			}

			if parsedDate.Year() == targetStartDate.Year() &&
				parsedDate.Month() == targetStartDate.Month() &&
				parsedDate.Day() == targetStartDate.Day() {
				startIdx = i
				break
			}
		}
		if startIdx == -1 {
			panic("Start date not found in CSV")
		}

		var endIdx int = -1
		for i, record := range records {
			if len(record) <= 0 {
				continue
			}
			parsedDate, err := parseDate(record, targetEndDate.Year())
			if err != nil {
				continue
			}

			if parsedDate.Year() == targetEndDate.Year() &&
				parsedDate.Month() == targetEndDate.Month() &&
				parsedDate.Day() == targetEndDate.Day() {
				endIdx = i
				break
			}
		}
		if endIdx == -1 {
			panic("End date not found in CSV")
		}

		logs.Log().Info(
			"Index",
			zap.Int("start", startIdx),
			zap.Int("end", endIdx),
		)

		var teacherRecords []ClassRecord
		currentDate := targetStartDate
		colIndices := parseColumnIndices()
		for i := startIdx + 1; i < len(records); i++ {
			record := records[i]
			if len(record) == 0 || strings.TrimSpace(record[0]) == "" {
				continue
			}

			parsedDate, err := parseDate(record, targetEndDate.Year())
			if err == nil {
				if parsedDate.After(targetEndDate) {
					break
				}
				currentDate = *parsedDate
				continue
			}

			if currentDate.After(targetEndDate) {
				break
			}

			teacherRec := parseTeacherRecord(record, currentDate, colIndices)
			if teacherRec != nil {
				teacherRecords = append(teacherRecords, *teacherRec)
			}
		}

		logs.Log().Info(
			"Parsed teacher records",
			zap.Int("count", len(teacherRecords)),
		)

		// for _, rec := range teacherRecords {
		// 	logs.Log().Info("Record",
		// 		zap.String("name", rec.Name),
		// 		zap.String("date", rec.Date),
		// 		zap.Int("timeMin", rec.TimeMin),
		// 		// zap.Timep("start time", rec.StartTime),
		// 		// zap.Timep("end time", rec.EndTime),
		// 		zap.Float64("rate", rec.Rate),
		// 		zap.String("status", rec.Status),
		// 	)
		// }

		sortedTeacherRecords := make([]ClassRecord, len(teacherRecords))
		copy(sortedTeacherRecords, teacherRecords)
		sort.Slice(sortedTeacherRecords, func(i, j int) bool {
			if sortedTeacherRecords[i].Name != sortedTeacherRecords[j].Name {
				return sortedTeacherRecords[i].Name < sortedTeacherRecords[j].Name
			}
			dateI, errI := parseDateString(sortedTeacherRecords[i].Date)
			dateJ, errJ := parseDateString(sortedTeacherRecords[j].Date)
			if errI != nil {
				return false
			}
			if errJ != nil {
				return true
			}
			return dateI.Before(*dateJ)
		})

		// TODO: (Brandon) Save to output file or process further
		for _, rec := range sortedTeacherRecords {
			logs.Log().Info("Record",
				zap.String("name", rec.Name),
				zap.String("date", rec.Date),
				zap.Int("timeMin", rec.DurationInMinutes),
				// zap.Timep("start time", rec.StartTime),
				// zap.Timep("end time", rec.EndTime),
				zap.Float64("rate", rec.Rate),
				zap.String("status", rec.Status),
			)
		}

		if err := saveRecordsToCSV(sortedTeacherRecords, parseTeacherFlags.output); err != nil {
			panic(err)
		}
	},
}

type ColumnIndices struct {
	Name      int
	Duration  int
	Rate      int
	Status    int
	StartTime int
	EndTime   int
	Link      int
}

func parseColumnIndices() ColumnIndices {
	return ColumnIndices{
		Name:      columnLetterToIndex(parseTeacherFlags.nameCol),
		Duration:  columnLetterToIndex(parseTeacherFlags.durationCol),
		Rate:      columnLetterToIndex(parseTeacherFlags.rateCol),
		Status:    columnLetterToIndex(parseTeacherFlags.statusCol),
		StartTime: columnLetterToIndex(parseTeacherFlags.startTimeCol),
		EndTime:   columnLetterToIndex(parseTeacherFlags.endTimeCol),
		Link:      columnLetterToIndex(parseTeacherFlags.linkCol),
	}
}

func columnLetterToIndex(col string) int {
	col = strings.ToUpper(strings.TrimSpace(col))
	if len(col) == 0 {
		return -1
	}
	return int(col[0] - 'A')
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
		parsedTime, err := parseClassTime(record[colIndices.StartTime], date)
		if err == nil {
			startTime = parsedTime
		}
	}

	// End time column
	var endTime *time.Time
	if colIndices.EndTime >= 0 && len(record) > colIndices.EndTime {
		parsedTime, err := parseClassTime(record[colIndices.EndTime], date)
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

func parseTimeString(timeStr string, date time.Time) (time.Time, error) {
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

func parseDateString(dateStr string) (*time.Time, error) {
	return parseDateStringWithYear(dateStr, 0)
}

func parseDateStringWithYear(dateStr string, year int) (*time.Time, error) {
	dateStr = strings.Trim(dateStr, `"`)
	dateStr = strings.TrimSpace(dateStr)

	// Normalize the date string: remove periods from month abbreviations
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

func parseDate(record []string, year int) (*time.Time, error) {
	if len(record) == 0 {
		return nil, errors.New("empty record")
	}
	return parseDateStringWithYear(record[0], year)
}

func parseClassTime(record string, date time.Time) (*time.Time, error) {
	timeStr := strings.TrimSpace(record)
	if timeStr != "" {
		parsedTime, err := parseTimeString(timeStr, date)
		if err != nil {
			return nil, err
		}
		return &parsedTime, nil
	}
	return nil, errors.New("invalid time")
}

func saveRecordsToCSV(records []ClassRecord, outputPath string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	colIndices := parseColumnIndices()
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
	if err := writer.Write([]string{fmt.Sprintf("TOTAL: %f", total)}); err != nil {
		return err
	}

	logs.Log().Info("Successfully saved CSV file", zap.String("output", outputPath), zap.Int("records", len(records)))
	return nil
}
