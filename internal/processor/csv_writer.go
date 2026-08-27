package processor

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

type CSVWriter struct{}

func (w *CSVWriter) WriteRecords(records []ClassRecord, outputPath string, colIndices ColumnIndices, name string, summary *RecordSummary) error {
	_ = summary
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	if name != "" {
		if err := writer.Write([]string{fmt.Sprintf("Teacher Name: %s", name)}); err != nil {
			return err
		}
		if err := writer.Write([]string{""}); err != nil {
			return err
		}
	}

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

func SaveRecordsToCSV(records []ClassRecord, outputPath string, colIndices ColumnIndices) error {
	writer := &CSVWriter{}
	return writer.WriteRecords(records, outputPath, colIndices, "", nil)
}
