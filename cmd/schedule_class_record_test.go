package cmd

import (
	"database/sql"
	"testing"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
)

func TestClassRecordRequestFromScheduleZeroesRateForCancelled(t *testing.T) {
	existing := queries.GetScheduledClassByIDRow{
		StudentID:       1,
		TeacherID:       2,
		ScheduledDate:   "2026-08-01",
		StartTime:       sql.NullString{String: "10:00", Valid: true},
		DurationMinutes: 60,
		Rate:            500,
		Currency:        "KRW",
	}

	req := classRecordRequestFromSchedule(existing, string(constants.ClassStatusCancelled), "student sick", "")
	if err := validateClassRecordRequest(&req); err != nil {
		t.Fatalf("validateClassRecordRequest() unexpected error: %v", err)
	}
	if req.Rate != 0 {
		t.Fatalf("req.Rate = %v, want 0", req.Rate)
	}
}
