package cmd

import (
	"sort"
	"strings"
	"zion-english/internal/logs"
	"zion-english/internal/processor"

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

var cmdParseTeacher = &cobra.Command{
	Use:   "parse_teacher",
	Short: "Parse teacher",
	Run: func(cmd *cobra.Command, args []string) {
		startDateStr := strings.TrimSpace(parseTeacherFlags.startDate)
		parsedStartDate, err := processor.ParseDateString(startDateStr)
		if err != nil {
			panic(err)
		}
		targetStartDate := *parsedStartDate

		endDateStr := strings.TrimSpace(parseTeacherFlags.endDate)
		parsedEndDate, err := processor.ParseDateString(endDateStr)
		if err != nil {
			panic(err)
		}
		targetEndDate := *parsedEndDate

		logs.Log().Info(
			"Processing teacher",
			zap.String("filepath", parseTeacherFlags.filepath),
			zap.String("output", parseTeacherFlags.output),
			zap.Time("start date", targetStartDate),
			zap.Time("end date", targetEndDate),
		)

		// Build column indices using processor functions
		colIndices := processor.ColumnIndices{
			Name:      processor.ColumnLetterToIndex(parseTeacherFlags.nameCol),
			Duration:  processor.ColumnLetterToIndex(parseTeacherFlags.durationCol),
			Rate:      processor.ColumnLetterToIndex(parseTeacherFlags.rateCol),
			Status:    processor.ColumnLetterToIndex(parseTeacherFlags.statusCol),
			StartTime: processor.ColumnLetterToIndex(parseTeacherFlags.startTimeCol),
			EndTime:   processor.ColumnLetterToIndex(parseTeacherFlags.endTimeCol),
			Link:      processor.ColumnLetterToIndex(parseTeacherFlags.linkCol),
		}

		teacherRecords, err := processor.ProcessCSVFile(parseTeacherFlags.filepath, targetStartDate, targetEndDate, colIndices, nil)
		if err != nil {
			panic(err)
		}

		logs.Log().Info(
			"Parsed teacher records",
			zap.Int("count", len(teacherRecords)),
		)

		var totalRate float64
		// for _, rec := range teacherRecords {
		// 	logs.Log().Info("Record",
		// 		zap.String("name", rec.Name),
		// 		zap.String("date", rec.Date),
		// 		zap.Int("duration (min)", rec.DurationInMinutes),
		// 		// zap.Timep("start time", rec.StartTime),
		// 		// zap.Timep("end time", rec.EndTime),
		// 		zap.Float64("rate", rec.Rate),
		// 		zap.String("status", rec.Status),
		// 	)
		// 	totalRate += rec.Rate
		// }
		logs.Log().Info(
			"Total rate",
			zap.Float64("total", totalRate),
		)

		sortedTeacherRecords := make([]processor.ClassRecord, len(teacherRecords))
		copy(sortedTeacherRecords, teacherRecords)
		sort.Slice(sortedTeacherRecords, func(i, j int) bool {
			if sortedTeacherRecords[i].Name != sortedTeacherRecords[j].Name {
				return sortedTeacherRecords[i].Name < sortedTeacherRecords[j].Name
			}
			dateI, errI := processor.ParseDateString(sortedTeacherRecords[i].Date)
			dateJ, errJ := processor.ParseDateString(sortedTeacherRecords[j].Date)
			if errI != nil {
				return false
			}
			if errJ != nil {
				return true
			}
			return dateI.Before(*dateJ)
		})

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
			totalRate += rec.Rate
		}
		logs.Log().Info(
			"Total rate",
			zap.Float64("total", totalRate),
		)

		if err := processor.SaveRecords(sortedTeacherRecords, parseTeacherFlags.output, colIndices, ""); err != nil {
			panic(err)
		}
	},
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

	rootCmd.AddCommand(cmdParseTeacher)
}
