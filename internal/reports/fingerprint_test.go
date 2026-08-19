package reports

import (
	"database/sql"
	"testing"
)

func TestFingerprintStableForSameRecords(t *testing.T) {
	rows := []FingerprintRow{
		{
			ID: 1, StudentID: 10, Date: "2026-08-01",
			StartTime: sql.NullString{String: "09:00", Valid: true},
			EndTime:   sql.NullString{String: "10:00", Valid: true},
			DurationMinutes: 60, Rate: 100.5, Currency: "KRW",
			Status: "conducted", UpdatedAt: "2026-08-01 10:00:00",
		},
		{
			ID: 2, StudentID: 11, Date: "2026-08-02",
			DurationMinutes: 45, Rate: 80, Currency: "KRW",
			Status: "cancelled", UpdatedAt: "2026-08-02 11:00:00",
		},
	}

	first := Fingerprint(rows)
	second := Fingerprint(rows)
	if first != second {
		t.Fatalf("expected stable fingerprint, got %q and %q", first, second)
	}
}

func TestFingerprintChangesWhenRecordChanges(t *testing.T) {
	base := FingerprintRow{
		ID: 1, StudentID: 10, Date: "2026-08-01",
		DurationMinutes: 60, Rate: 100, Currency: "KRW",
		Status: "conducted", UpdatedAt: "2026-08-01 10:00:00",
	}
	changed := base
	changed.Rate = 120

	if Fingerprint([]FingerprintRow{base}) == Fingerprint([]FingerprintRow{changed}) {
		t.Fatal("expected different fingerprint when rate changes")
	}
}

func TestFingerprintEmptySet(t *testing.T) {
	empty := Fingerprint(nil)
	if empty == "" {
		t.Fatal("expected non-empty hash for empty record set")
	}
	if Fingerprint(nil) != empty {
		t.Fatal("expected stable hash for empty record set")
	}
}
