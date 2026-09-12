package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestSaveSummaryReport(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "summary.xlsx")

	sheets := []SummaryTeacherSheet{
		{
			TeacherName: "Don Teacher",
			Students: []SummaryStudentBlock{
				{
					StudentName: "Student One",
					Classes: []SummaryClassRow{
						{DateDisplay: "August 1", RateDisplay: "20.00 KRW", RateValue: 20, HasRate: true},
						{DateDisplay: "August 2", RateDisplay: "20.00 KRW", RateValue: 20, HasRate: true},
					},
					Total:    40,
					Currency: "KRW",
					HasTotal: true,
				},
			},
		},
	}

	if err := SaveSummaryReport(sheets, outputPath); err != nil {
		t.Fatalf("SaveSummaryReport: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file missing: %v", err)
	}

	f, err := excelize.OpenFile(outputPath)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName != "Don Teacher" {
		t.Fatalf("sheet name = %q, want %q", sheetName, "Don Teacher")
	}
	if got, _ := f.GetCellValue(sheetName, "A1"); got != "Don Teacher" {
		t.Fatalf("A1 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "A2"); got != "Student One" {
		t.Fatalf("A2 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "A3"); got != "August 1" {
		t.Fatalf("A3 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "B3"); got != "20.00 KRW" {
		t.Fatalf("B3 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "A5"); got != "total in krw" {
		t.Fatalf("A5 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "B5"); got != "40" {
		t.Fatalf("B5 = %q", got)
	}
}
