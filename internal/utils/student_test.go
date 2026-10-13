package utils

import (
	"database/sql"
	"testing"
)

func TestStudentDisplayName(t *testing.T) {
	if got := StudentDisplayName("Suhyun", "Z001"); got != "Suhyun (Z001)" {
		t.Fatalf("got %q", got)
	}
	if got := StudentDisplayName("Suhyun", ""); got != "Suhyun" {
		t.Fatalf("got %q", got)
	}
	if got := StudentDisplayNameFromNull("Lee", sql.NullString{String: "Z032", Valid: true}); got != "Lee (Z032)" {
		t.Fatalf("got %q", got)
	}
}
