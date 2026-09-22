package cmd

import (
	"database/sql"
	"testing"
	"time"
	"zion-english/internal/database/queries"
)

func TestTeacherRowsTiebreakUsesBothLastNames(t *testing.T) {
	created := mustParseTime("2024-01-01T10:00:00Z")
	a := queries.GetTeachersFilteredRow{
		FirstName: "Amy",
		LastName:  "Zulu",
		CreatedAt: created,
	}
	b := queries.GetTeachersFilteredRow{
		FirstName: "Amy",
		LastName:  "Alpha",
		CreatedAt: created,
	}

	if teacherRowsTiebreak(a, b) >= 0 {
		t.Fatalf("expected Alpha before Zulu when created_at ties, got tiebreak >= 0")
	}
	if teacherRowsTiebreak(b, a) <= 0 {
		t.Fatalf("expected Zulu after Alpha when created_at ties, got tiebreak <= 0")
	}
}

func mustParseTime(value string) sql.NullTime {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return sql.NullTime{Valid: true, Time: t}
}
