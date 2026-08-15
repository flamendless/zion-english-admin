package utils

import (
	"database/sql"
	"testing"
	"time"
)

func TestPersonInitials(t *testing.T) {
	tests := []struct {
		first, middle, last, full string
		want                      string
	}{
		{"Alice", "", "Smith", "", "AS"},
		{"Alice", "Marie", "Smith", "", "AS"},
		{"Alice", "", "", "", "A"},
		{"", "", "", "John Doe", "JD"},
		{"", "", "", "Cher", "CH"},
		{"", "", "", "", "?"},
	}

	for _, tt := range tests {
		got := PersonInitials(tt.first, tt.middle, tt.last, tt.full)
		if got != tt.want {
			t.Fatalf("PersonInitials(%q, %q, %q, %q) = %q, want %q", tt.first, tt.middle, tt.last, tt.full, got, tt.want)
		}
	}
}

func TestSensitiveChangeAllowed(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)

	allowed, days := SensitiveChangeAllowed(sql.NullTime{}, now)
	if !allowed || days != 0 {
		t.Fatalf("expected allowed with no prior change, got allowed=%v days=%d", allowed, days)
	}

	recent := sql.NullTime{Time: now.Add(-24 * time.Hour), Valid: true}
	allowed, days = SensitiveChangeAllowed(recent, now)
	if allowed || days != 6 {
		t.Fatalf("expected locked with 6 days remaining, got allowed=%v days=%d", allowed, days)
	}

	old := sql.NullTime{Time: now.Add(-8 * 24 * time.Hour), Valid: true}
	allowed, days = SensitiveChangeAllowed(old, now)
	if !allowed || days != 0 {
		t.Fatalf("expected allowed after cooldown, got allowed=%v days=%d", allowed, days)
	}
}
