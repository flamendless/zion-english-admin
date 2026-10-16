package processor

import (
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestColWidthTrackerApply(t *testing.T) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	tracker := newColWidthTracker(2, 10, []float64{28, 48})
	tracker.Note(1, "Conducted")
	tracker.Note(2, "Very Long Student Name For Testing Purposes")
	if err := tracker.Apply(f, sheet); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	wB, err := f.GetColWidth(sheet, "B")
	if err != nil {
		t.Fatalf("GetColWidth: %v", err)
	}
	if wB < 20 {
		t.Fatalf("col B width %v, expected wider for long name", wB)
	}
}
