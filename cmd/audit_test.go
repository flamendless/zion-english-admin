package cmd

import (
	"database/sql"
	"strings"
	"testing"
	"zion-english/internal/database/queries"
)

func TestFormatTeacherAudit_fieldChanges(t *testing.T) {
	before := queries.GetTeacherFullByIDRow{
		ID:             5,
		FirstName:      "Jane",
		LastName:       "Doe",
		Email:          "old@example.com",
		Template:       sql.NullString{String: "A,B,C,G", Valid: true},
		RatePerClass:   10,
		Certifications: sql.NullString{Valid: false},
	}
	after := queries.GetTeacherFullByIDRow{
		ID:             5,
		FirstName:      "Jane",
		LastName:       "Doe",
		Email:          "new@example.com",
		Template:       sql.NullString{String: "A,C,D,H", Valid: true},
		RatePerClass:   12,
		Certifications: sql.NullString{Valid: false},
	}

	msg := formatTeacherAudit(before, after)
	for _, want := range []string{
		"email 'old@example.com' -> 'new@example.com'",
		"template 'A,B,C,G' -> 'A,C,D,H'",
		"rate 10.00 -> 12.00",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q: %s", want, msg)
		}
	}
}

func TestFormatTeacherAudit_noChanges(t *testing.T) {
	row := queries.GetTeacherFullByIDRow{ID: 3, FirstName: "Bob", LastName: "Smith"}
	msg := formatTeacherAudit(row, row)
	if !strings.Contains(msg, "(no field changes)") {
		t.Errorf("expected no-change message, got: %s", msg)
	}
}
