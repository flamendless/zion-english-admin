package database

import (
	"testing"
	"time"

	"zion-english/internal/constants"
)

func TestMonthlyPlanEffectiveEndPHT(t *testing.T) {
	end, err := MonthlyPlanEffectiveEndPHT("2026-02-15")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 2, 28, 23, 59, 59, 0, constants.LocationPHT)
	gotEnd, _ := time.ParseInLocation(constants.DateTimeSecondsLayout, end, constants.LocationPHT)
	if gotEnd.Year() != want.Year() || gotEnd.Month() != want.Month() || gotEnd.Day() != want.Day() {
		t.Fatalf("got %q want end of Feb 2026 PHT", end)
	}
}

func TestCurrentMonthDateRangePHT(t *testing.T) {
	start, end := CurrentMonthDateRangePHT()
	if start == "" || end == "" {
		t.Fatal("expected non-empty range")
	}
	if len(start) != 10 || len(end) != 10 {
		t.Fatalf("expected YYYY-MM-DD dates, got %q %q", start, end)
	}
}
