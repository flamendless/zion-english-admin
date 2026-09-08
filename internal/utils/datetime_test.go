package utils

import (
	"database/sql"
	"testing"
	"time"

	"zion-english/internal/constants"
)

func TestDateTimePHT(t *testing.T) {
	utc := time.Date(2026, 8, 20, 6, 30, 0, 0, time.UTC)
	got := DateTimePHT(utc)
	want := "2026-08-20 14:30"
	if got != want {
		t.Fatalf("DateTimePHT() = %q, want %q", got, want)
	}
}

func TestFormatNullDateTimePHT(t *testing.T) {
	if got := FormatNullDateTimePHT(sql.NullTime{}); got != "-" {
		t.Fatalf("invalid null = %q, want %q", got, "-")
	}

	utc := time.Date(2026, 1, 1, 16, 0, 0, 0, time.UTC)
	got := FormatNullDateTimePHT(sql.NullTime{Time: utc, Valid: true})
	if got != "2026-01-02 00:00" {
		t.Fatalf("FormatNullDateTimePHT() = %q, want midnight PHT next day", got)
	}
}

func TestFormatCompactDateRanges(t *testing.T) {
	tests := []struct {
		name  string
		dates []string
		want  string
	}{
		{
			name:  "empty",
			dates: nil,
			want:  "-",
		},
		{
			name:  "single date",
			dates: []string{"2026-09-14"},
			want:  "Sep 14, 2026",
		},
		{
			name:  "contiguous range",
			dates: []string{"2026-09-14", "2026-09-15", "2026-09-16"},
			want:  "Sep 14 - 16, 2026",
		},
		{
			name:  "gapped dates",
			dates: []string{"2026-09-14", "2026-09-15", "2026-09-17", "2026-09-18", "2026-09-19"},
			want:  "Sep 14 - 15, 17 - 19, 2026",
		},
		{
			name:  "same month non consecutive singles",
			dates: []string{"2026-09-14", "2026-09-19"},
			want:  "Sep 14, 19, 2026",
		},
		{
			name:  "cross month consecutive",
			dates: []string{"2026-09-30", "2026-10-01"},
			want:  "Sep 30 - Oct 1, 2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatCompactDateRanges(tt.dates); got != tt.want {
				t.Fatalf("FormatCompactDateRanges() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTodayPHTUsesManila(t *testing.T) {
	now := time.Now().In(constants.LocationPHT)
	if TodayPHT() != now.Format(constants.DateLayout) {
		t.Fatalf("TodayPHT() should match current PHT calendar date")
	}
}
