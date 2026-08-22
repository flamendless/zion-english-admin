package utils

import (
	"strings"
	"testing"
)

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
