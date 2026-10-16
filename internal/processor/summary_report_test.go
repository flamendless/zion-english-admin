package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
	"zion-english/internal/constants"
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
						{
							Status:           constants.ClassStatusConducted,
							StatusDisplay:    "Conducted",
							DateDisplay:      "August 1",
							RateDisplay:      "20.00 KRW",
							RateValue:        20,
							HasRate:          true,
							ClassRateDisplay: "15.00 KRW",
						},
						{
							Status:           constants.ClassStatusCancelled,
							StatusDisplay:    "Cancelled",
							DateDisplay:      "August 2",
							RateDisplay:      "10.00 KRW",
							RateValue:        10,
							HasRate:          true,
							ClassRateDisplay: "0.00 KRW",
						},
					},
					Total:    30,
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
	if got, _ := f.GetCellValue(sheetName, "B2"); got != "Student One" {
		t.Fatalf("B2 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "A3"); got != "Conducted" {
		t.Fatalf("A3 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "B3"); got != "August 1" {
		t.Fatalf("B3 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "C3"); got != "20.00 KRW" {
		t.Fatalf("C3 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "D3"); got != "15.00 KRW" {
		t.Fatalf("D3 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "B5"); got != "total in krw" {
		t.Fatalf("B5 = %q", got)
	}
	if got, _ := f.GetCellValue(sheetName, "C5"); got != "30" {
		t.Fatalf("C5 = %q", got)
	}
}
