package processor

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

type SummaryClassRow struct {
	Status           constants.ClassStatus
	StatusDisplay    string
	DateDisplay      string
	RateDisplay      string
	RateValue        float64
	HasRate          bool
	ClassRateDisplay string
}

type SummaryStudentBlock struct {
	StudentName string
	Classes     []SummaryClassRow
	Total       float64
	Currency    string
	HasTotal    bool
}

type SummaryTeacherSheet struct {
	TeacherName string
	Students    []SummaryStudentBlock
}

type summaryRowStyles struct {
	status int
	date   int
	rate   int
}

func summaryStatusFillColor(status constants.ClassStatus) string {
	switch status {
	case constants.ClassStatusConducted:
		return "ECFDF5"
	case constants.ClassStatusCancelled:
		return "FEE2E2"
	case constants.ClassStatusRescheduled:
		return "FFFBEB"
	default:
		return "FFFFFF"
	}
}

func newSummaryRowStyles(f *excelize.File, status constants.ClassStatus) (summaryRowStyles, error) {
	fill := summaryStatusFillColor(status)
	baseFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{fill}}
	statusStyle, err := f.NewStyle(&excelize.Style{Fill: baseFill})
	if err != nil {
		return summaryRowStyles{}, err
	}
	dateStyle, err := f.NewStyle(&excelize.Style{
		Fill:      baseFill,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	if err != nil {
		return summaryRowStyles{}, err
	}
	rateStyle, err := f.NewStyle(&excelize.Style{Fill: baseFill})
	if err != nil {
		return summaryRowStyles{}, err
	}
	return summaryRowStyles{status: statusStyle, date: dateStyle, rate: rateStyle}, nil
}

func SaveSummaryReport(sheets []SummaryTeacherSheet, outputPath string) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	if len(sheets) == 0 {
		return fmt.Errorf("no sheets to write")
	}

	usedNames := map[string]int{}
	boldStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return err
	}

	rowStyleCache := map[constants.ClassStatus]summaryRowStyles{}

	for i, sheet := range sheets {
		sheetName := uniqueSheetName(sheet.TeacherName, usedNames)
		var sheetIndex int
		if i == 0 {
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return err
			}
			sheetIndex = 0
		} else {
			sheetIndex, err = f.NewSheet(sheetName)
			if err != nil {
				return err
			}
		}
		f.SetActiveSheet(sheetIndex)

		colWidths := newColWidthTracker(4, 10, []float64{28, 48, 28, 28})

		rowNum := 1
		cell, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetCellValue(sheetName, cell, sheet.TeacherName)
		colWidths.Note(1, sheet.TeacherName)
		f.SetCellStyle(sheetName, cell, cell, boldStyle)
		rowNum++

		for _, student := range sheet.Students {
			cell, _ = excelize.CoordinatesToCellName(2, rowNum)
			f.SetCellValue(sheetName, cell, student.StudentName)
			colWidths.Note(2, student.StudentName)
			f.SetCellStyle(sheetName, cell, cell, boldStyle)
			rowNum++

			for _, classRow := range student.Classes {
				styles, ok := rowStyleCache[classRow.Status]
				if !ok {
					styles, err = newSummaryRowStyles(f, classRow.Status)
					if err != nil {
						return err
					}
					rowStyleCache[classRow.Status] = styles
				}

				cellA, _ := excelize.CoordinatesToCellName(1, rowNum)
				f.SetCellValue(sheetName, cellA, classRow.StatusDisplay)
				colWidths.Note(1, classRow.StatusDisplay)
				f.SetCellStyle(sheetName, cellA, cellA, styles.status)

				cellB, _ := excelize.CoordinatesToCellName(2, rowNum)
				f.SetCellValue(sheetName, cellB, classRow.DateDisplay)
				colWidths.Note(2, classRow.DateDisplay)
				f.SetCellStyle(sheetName, cellB, cellB, styles.date)

				cellC, _ := excelize.CoordinatesToCellName(3, rowNum)
				f.SetCellValue(sheetName, cellC, classRow.RateDisplay)
				colWidths.Note(3, classRow.RateDisplay)
				f.SetCellStyle(sheetName, cellC, cellC, styles.rate)

				cellD, _ := excelize.CoordinatesToCellName(4, rowNum)
				f.SetCellValue(sheetName, cellD, classRow.ClassRateDisplay)
				colWidths.Note(4, classRow.ClassRateDisplay)
				f.SetCellStyle(sheetName, cellD, cellD, styles.rate)
				rowNum++
			}

			if student.HasTotal {
				totalLabel := fmt.Sprintf("total in %s", strings.ToLower(student.Currency))
				cellB, _ := excelize.CoordinatesToCellName(2, rowNum)
				f.SetCellValue(sheetName, cellB, totalLabel)
				colWidths.Note(2, totalLabel)
				cellC, _ := excelize.CoordinatesToCellName(3, rowNum)
				f.SetCellValue(sheetName, cellC, student.Total)
				colWidths.NoteFloat(3, student.Total)
				rowNum++
			}

			rowNum++
		}

		if err := colWidths.Apply(f, sheetName); err != nil {
			return err
		}
	}

	return f.SaveAs(outputPath)
}

func uniqueSheetName(name string, used map[string]int) string {
	base := utils.SanitizeSheetName(name)
	if base == "" {
		base = "Sheet"
	}
	candidate := base
	for {
		if _, exists := used[candidate]; !exists {
			used[candidate] = 1
			return candidate
		}
		used[candidate]++
		suffix := fmt.Sprintf(" (%d)", used[candidate])
		maxBase := 31 - len(suffix)
		if maxBase < 1 {
			maxBase = 1
		}
		if len(base) > maxBase {
			candidate = base[:maxBase] + suffix
		} else {
			candidate = base + suffix
		}
	}
}
