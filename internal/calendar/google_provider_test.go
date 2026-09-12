package calendar

import (
	"testing"

	"zion-english/internal/constants"
)

func TestEventDateTimes(t *testing.T) {
	start, end, err := eventDateTimes("2026-08-29", "10:00", 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start == "" || end == "" {
		t.Fatal("expected non-empty start and end")
	}
	if want := constants.TimezoneNamePHT; want == "" {
		t.Fatal("timezone constant missing")
	}
}

func TestEventDateTimesRequiresStartTime(t *testing.T) {
	_, _, err := eventDateTimes("2026-08-29", "", 60)
	if err != ErrStartTimeRequired {
		t.Fatalf("expected ErrStartTimeRequired, got %v", err)
	}
}

func TestBuildEventDescription(t *testing.T) {
	desc := buildEventDescription(ScheduledClassCalendarInput{
		StudentName:     "Alice",
		ScheduledDate:   "2026-08-29",
		StartTime:       "10:00",
		DurationMinutes: 60,
		Rate:            500,
		Currency:        "PHP",
	})
	if desc == "" {
		t.Fatal("expected description")
	}
}
