package utils

import (
	"errors"
	"testing"
)

func TestDurationMinutesFromRange(t *testing.T) {
	d, err := DurationMinutesFromRange("10:00", "11:30")
	if err != nil {
		t.Fatal(err)
	}
	if d != 90 {
		t.Fatalf("expected 90, got %d", d)
	}
}

func TestDurationMinutesFromRangeInvalid(t *testing.T) {
	_, err := DurationMinutesFromRange("11:00", "10:00")
	if !errors.Is(err, ErrEndBeforeStart) {
		t.Fatalf("expected ErrEndBeforeStart, got %v", err)
	}
}

func TestParseTimeHMRequired(t *testing.T) {
	_, err := ParseTimeHM("")
	if !errors.Is(err, ErrTimeRequired) {
		t.Fatalf("expected ErrTimeRequired, got %v", err)
	}
}

func TestDurationMinutesFromRangeInvalidStart(t *testing.T) {
	_, err := DurationMinutesFromRange("bad", "10:00")
	if !errors.Is(err, ErrInvalidStartTime) {
		t.Fatalf("expected ErrInvalidStartTime, got %v", err)
	}
}

func TestEndTimeFromStartAndDuration(t *testing.T) {
	got := EndTimeFromStartAndDuration("09:15", 45)
	if got != "10:00" {
		t.Fatalf("expected 10:00, got %s", got)
	}
}
