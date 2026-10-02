package announcements

import (
	"database/sql"
	"testing"
)

func TestIsVisibleOnDate(t *testing.T) {
	base := VisibilityInput{
		StartDate: "2026-09-01",
		EndDate:   "2026-09-30",
	}
	if !IsVisibleOnDate(base, "2026-09-10") {
		t.Fatal("expected visible without repeat")
	}
	if IsVisibleOnDate(base, "2026-08-31") {
		t.Fatal("expected before start to be hidden")
	}

	withCutoff := VisibilityInput{
		StartDate:        "2026-09-01",
		EndDate:          "2026-09-30",
		RepeatEnabled:    true,
		RepeatSchedule:   RepeatScheduleCutoffBefore,
		CutoffRepeatDays: sql.NullInt64{Int64: 3, Valid: true},
	}
	// Depends on ActiveCutoffDates at runtime; use a day outside campaign to ensure date gate works.
	if IsVisibleOnDate(withCutoff, "2026-10-01") {
		t.Fatal("expected after end to be hidden")
	}
}
