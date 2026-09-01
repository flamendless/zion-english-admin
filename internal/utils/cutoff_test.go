package utils

import (
	"strings"
	"testing"
	"time"
)

func TestCutoffRangeForMonth(t *testing.T) {
	first, second := CutoffRangeForMonth(2026, time.September)
	if first != "2026-09-01|2026-09-15" {
		t.Fatalf("unexpected first cutoff: %q", first)
	}
	if second != "2026-09-16|2026-09-30" {
		t.Fatalf("unexpected second cutoff: %q", second)
	}
}

func TestParseMonthPHT(t *testing.T) {
	year, month, ok := ParseMonthPHT("2026-09")
	if !ok || year != 2026 || month != time.September {
		t.Fatalf("unexpected parse result: %d %v %v", year, month, ok)
	}
	if _, _, ok := ParseMonthPHT("invalid"); ok {
		t.Fatal("expected invalid month to fail")
	}
}

func TestActiveCutoffDates(t *testing.T) {
	start, end := ActiveCutoffDates()
	if start == "" || end == "" {
		t.Fatal("expected non-empty cutoff dates")
	}
	if strings.Count(start, "-") != 2 || strings.Count(end, "-") != 2 {
		t.Fatalf("unexpected date format: %q %q", start, end)
	}
	if start > end {
		t.Fatalf("start %q after end %q", start, end)
	}
}

func TestCurrentCutoffRange(t *testing.T) {
	first, second, active := CurrentCutoffRange()
	if first == "" || second == "" || active == "" {
		t.Fatal("expected cutoff presets")
	}
	if active != first && active != second {
		t.Fatalf("active cutoff %q must match first %q or second %q", active, first, second)
	}
}
