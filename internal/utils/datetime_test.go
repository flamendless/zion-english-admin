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

func TestTodayPHTUsesManila(t *testing.T) {
	now := time.Now().In(constants.LocationPHT)
	if TodayPHT() != now.Format(constants.DateLayout) {
		t.Fatalf("TodayPHT() should match current PHT calendar date")
	}
}
