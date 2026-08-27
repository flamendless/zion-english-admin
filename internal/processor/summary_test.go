package processor

import (
	"testing"

	"zion-english/internal/constants"
)

func TestComputeRecordSummary(t *testing.T) {
	records := []ClassRecord{
		{Status: string(constants.ClassStatusConducted), DurationInMinutes: 60},
		{Status: string(constants.ClassStatusConducted), DurationInMinutes: 45},
		{Status: string(constants.ClassStatusCancelled), DurationInMinutes: 30},
		{Status: string(constants.ClassStatusRescheduled), DurationInMinutes: 60},
	}

	summary := ComputeRecordSummary(records)

	if summary.TotalClasses != 4 {
		t.Fatalf("TotalClasses = %d, want 4", summary.TotalClasses)
	}
	if summary.TotalConducted != 2 {
		t.Fatalf("TotalConducted = %d, want 2", summary.TotalConducted)
	}
	if summary.TotalCancelled != 1 {
		t.Fatalf("TotalCancelled = %d, want 1", summary.TotalCancelled)
	}
	if summary.TotalDuration != 195 {
		t.Fatalf("TotalDuration = %d, want 195", summary.TotalDuration)
	}
	if summary.DurationDisplay != "3 hr 15 min" {
		t.Fatalf("DurationDisplay = %q, want %q", summary.DurationDisplay, "3 hr 15 min")
	}
}
