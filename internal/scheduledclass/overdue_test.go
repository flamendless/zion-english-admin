package scheduledclass_test

import (
	"testing"
	"time"
	"zion-english/internal/constants"
	"zion-english/internal/scheduledclass"
)

func TestIsOverdueBeforeGraceEnds(t *testing.T) {
	now := time.Date(2026, 9, 12, 10, 5, 0, 0, constants.LocationPHT)
	overdue := scheduledclass.IsOverdue(
		constants.ScheduledClassStatusScheduled,
		"2026-09-12",
		"09:00",
		60,
		now,
		15,
	)
	if overdue {
		t.Fatal("expected class to stay pending during grace period")
	}
}

func TestIsOverdueAfterGraceEnds(t *testing.T) {
	now := time.Date(2026, 9, 12, 10, 16, 0, 0, constants.LocationPHT)
	overdue := scheduledclass.IsOverdue(
		constants.ScheduledClassStatusScheduled,
		"2026-09-12",
		"09:00",
		60,
		now,
		15,
	)
	if !overdue {
		t.Fatal("expected class to be overdue after grace period")
	}
}

func TestIsOverdueIgnoresNonScheduledStatus(t *testing.T) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, constants.LocationPHT)
	overdue := scheduledclass.IsOverdue(
		constants.ScheduledClassStatus("conducted"),
		"2026-09-12",
		"09:00",
		60,
		now,
		0,
	)
	if overdue {
		t.Fatal("expected conducted class not to be overdue")
	}
}

func TestIsOverdueAfterMidnightRollover(t *testing.T) {
	endAt, ok := scheduledclass.ScheduledEndAtPHT("2026-09-12", "23:30", 90)
	if !ok {
		t.Fatal("expected scheduled end time")
	}
	expected := time.Date(2026, 9, 13, 1, 0, 0, 0, constants.LocationPHT)
	if !endAt.Equal(expected) {
		t.Fatalf("expected end at %s, got %s", expected, endAt)
	}

	now := time.Date(2026, 9, 13, 1, 1, 0, 0, constants.LocationPHT)
	overdue := scheduledclass.IsOverdue(
		constants.ScheduledClassStatusScheduled,
		"2026-09-12",
		"23:30",
		90,
		now,
		0,
	)
	if !overdue {
		t.Fatal("expected class to be overdue after midnight rollover")
	}
}
