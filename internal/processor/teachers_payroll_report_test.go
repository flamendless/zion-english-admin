package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestSaveTeachersPayrollReport(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "payroll.xlsx")

	rows := []TeachersPayrollRow{
		{
			TeacherFullName: "Foo Bar",
			CutoffLabel:     "October 1 - 15",
			TotalClasses:    36,
			EarningsDisplay: "12,345.00 PHP",
		},
	}

	if err := SaveTeachersPayrollReport(rows, outputPath); err != nil {
		t.Fatalf("SaveTeachersPayrollReport: %v", err)
	}
	f, err := excelize.OpenFile(outputPath)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if got, _ := f.GetCellValue(sheet, "A1"); got != "Teacher" {
		t.Fatalf("A1 = %q", got)
	}
	if got, _ := f.GetCellValue(sheet, "A2"); got != "Foo Bar" {
		t.Fatalf("A2 = %q", got)
	}
	if got, _ := f.GetCellValue(sheet, "B2"); got != "October 1 - 15" {
		t.Fatalf("B2 = %q", got)
	}
	if got, _ := f.GetCellValue(sheet, "C2"); got != "36" {
		t.Fatalf("C2 = %q", got)
	}
	if got, _ := f.GetCellValue(sheet, "D2"); got != "12,345.00 PHP" {
		t.Fatalf("D2 = %q", got)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output missing: %v", err)
	}
}
