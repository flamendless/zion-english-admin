package processor

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type XLSXWriter struct{}

func (w *XLSXWriter) WriteRecords(records []ClassRecord, outputPath string, colIndices ColumnIndices, name string, summary *RecordSummary) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	sheetName := "Sheet1"
	f.SetActiveSheet(0)

	rowNum := 1

	if name != "" {
		cell, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetCellValue(sheetName, cell, fmt.Sprintf("Teacher Name: %s", name))
		rowNum++
		if summary != nil {
			summaryRows := []string{
				fmt.Sprintf("Total Classes: %d", summary.TotalClasses),
				fmt.Sprintf("Total Conducted: %d", summary.TotalConducted),
				fmt.Sprintf("Total Cancelled: %d", summary.TotalCancelled),
				fmt.Sprintf("Total Duration: %s", summary.DurationDisplay),
			}
			for _, line := range summaryRows {
				cell, _ = excelize.CoordinatesToCellName(1, rowNum)
				f.SetCellValue(sheetName, cell, line)
				rowNum++
			}
		}
		rowNum++
	}

	headers := []string{"Name", "Date", "Duration", "Rate"}
	colIndex := 1
	for _, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, header)
		colIndex++
	}

	if colIndices.StartTime >= 0 {
		cell, _ := excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, "StartTime")
		colIndex++
	}
	if colIndices.EndTime >= 0 {
		cell, _ := excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, "EndTime")
		colIndex++
	}
	if colIndices.Link >= 0 {
		cell, _ := excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, "GoogleLink")
		colIndex++
	}
	cell, _ := excelize.CoordinatesToCellName(colIndex, rowNum)
	f.SetCellValue(sheetName, cell, "Status")

	// Style header row
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4CAF50"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	headerRowCell := fmt.Sprintf("A%d", rowNum)
	f.SetCellStyle(sheetName, headerRowCell, getColumnLetter(colIndex)+fmt.Sprintf("%d", rowNum), headerStyle)

	rowNum++
	var total float64
	nameColorMap := make(map[string]string)
	uniqueNames := make([]string, 0)
	seenNames := make(map[string]bool)

	// First pass: collect unique names in order
	for _, rec := range records {
		if !seenNames[rec.Name] {
			uniqueNames = append(uniqueNames, rec.Name)
			seenNames[rec.Name] = true
		}
	}

	// Assign colors sequentially to ensure different colors for consecutive names
	for i, name := range uniqueNames {
		nameColorMap[name] = StudentColor(i)
	}

	for _, rec := range records {
		colIndex := 1

		// Get color for this name (already assigned)
		color := nameColorMap[rec.Name]

		// Name
		cell, _ := excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, rec.Name)
		colIndex++

		// Date
		cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, rec.Date)
		colIndex++

		// Duration
		cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, rec.DurationInMinutes)
		colIndex++

		// Rate
		cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, rec.Rate)
		colIndex++

		// StartTime
		if colIndices.StartTime >= 0 {
			cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
			if rec.StartTime != nil {
				f.SetCellValue(sheetName, cell, rec.StartTime.Format("15:04"))
			}
			colIndex++
		}

		// EndTime
		if colIndices.EndTime >= 0 {
			cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
			if rec.EndTime != nil {
				f.SetCellValue(sheetName, cell, rec.EndTime.Format("15:04"))
			}
			colIndex++
		}

		// GoogleLink
		if colIndices.Link >= 0 {
			cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
			f.SetCellValue(sheetName, cell, rec.GoogleLink)
			colIndex++
		}

		// Status
		cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
		f.SetCellValue(sheetName, cell, rec.Status)

		// Apply color to entire row
		rowStyle, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{color}, Pattern: 1},
		})
		if err != nil {
			return err
		}
		lastCol := getColumnLetter(colIndex)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("%s%d", lastCol, rowNum), rowStyle)

		total += rec.Rate
		rowNum++
	}

	// Add total row
	colIndex = 1
	cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
	f.SetCellValue(sheetName, cell, "TOTAL")
	colIndex += 3
	cell, _ = excelize.CoordinatesToCellName(colIndex, rowNum)
	f.SetCellValue(sheetName, cell, total)

	// Style total row
	totalStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F0F0F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	lastCol := getColumnLetter(colIndex)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("%s%d", lastCol, rowNum), totalStyle)

	return f.SaveAs(outputPath)
}

func getColumnLetter(colIndex int) string {
	result := ""
	for colIndex > 0 {
		colIndex--
		result = string(rune('A'+colIndex%26)) + result
		colIndex /= 26
	}
	return result
}
