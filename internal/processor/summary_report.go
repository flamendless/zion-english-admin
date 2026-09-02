package processor

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
	"zion-english/internal/utils"
)

type SummaryClassRow struct {
	DateDisplay string
	RateDisplay string
	RateValue   float64
	HasRate     bool
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

func SaveSummaryReport(sheets []SummaryTeacherSheet, outputPath string) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	if len(sheets) == 0 {
		return fmt.Errorf("no sheets to write")
	}

	usedNames := map[string]int{}
	rightAlignStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	boldStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return err
	}

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

		rowNum := 1
		cell, _ := excelize.CoordinatesToCellName(1, rowNum)
		f.SetCellValue(sheetName, cell, sheet.TeacherName)
		f.SetCellStyle(sheetName, cell, cell, boldStyle)
		rowNum++

		for _, student := range sheet.Students {
			cell, _ = excelize.CoordinatesToCellName(1, rowNum)
			f.SetCellValue(sheetName, cell, student.StudentName)
			f.SetCellStyle(sheetName, cell, cell, boldStyle)
			rowNum++

			for _, classRow := range student.Classes {
				cellA, _ := excelize.CoordinatesToCellName(1, rowNum)
				f.SetCellValue(sheetName, cellA, classRow.DateDisplay)
				f.SetCellStyle(sheetName, cellA, cellA, rightAlignStyle)

				cellB, _ := excelize.CoordinatesToCellName(2, rowNum)
				f.SetCellValue(sheetName, cellB, classRow.RateDisplay)
				rowNum++
			}

			if student.HasTotal {
				cellA, _ := excelize.CoordinatesToCellName(1, rowNum)
				f.SetCellValue(sheetName, cellA, fmt.Sprintf("total in %s", strings.ToLower(student.Currency)))
				cellB, _ := excelize.CoordinatesToCellName(2, rowNum)
				f.SetCellValue(sheetName, cellB, student.Total)
				rowNum++
			}

			rowNum++
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
