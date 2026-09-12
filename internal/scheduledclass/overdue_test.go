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
