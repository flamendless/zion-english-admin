package cmd

import (
	"database/sql"
	"testing"
)

func TestAuditStudentRef(t *testing.T) {
	got := auditStudentRef(31, "Huyoon", sql.NullString{})
	if got != "student id 31 (Huyoon)" {
		t.Fatalf("got %q", got)
	}
	got = auditStudentRef(31, "Lee", sql.NullString{String: "Z032", Valid: true})
	if got != "student id 31 (Lee (Z032))" {
		t.Fatalf("got %q", got)
	}
}

func TestAuditTeacherRef(t *testing.T) {
	if auditTeacherRef(38, "Brandon") != "teacher id 38 (Brandon)" {
		t.Fatal("named teacher")
	}
	if auditTeacherRef(38, "  ") != "teacher id 38 (unknown)" {
		t.Fatal("empty teacher name")
	}
}

func TestAuditStudentTeacherPair(t *testing.T) {
	got := auditStudentTeacherPair(31, "Huyoon", sql.NullString{}, 38, "Brandon")
	want := "student id 31 (Huyoon), teacher id 38 (Brandon)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
