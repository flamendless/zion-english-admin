package constants

import "testing"

func TestValidCurrency(t *testing.T) {
	for _, c := range []string{"KRW", "CAD", "YEN", "PHP"} {
		if !ValidCurrency(c) {
			t.Fatalf("ValidCurrency(%q) = false, want true", c)
		}
	}
	if ValidCurrency("USD") {
		t.Fatal("ValidCurrency(USD) = true, want false")
	}
}

func TestValidStudentStatus(t *testing.T) {
	if !ValidStudentStatus("active") || !ValidStudentStatus("inactive") {
		t.Fatal("expected active/inactive to be valid")
	}
	if ValidStudentStatus("pending") {
		t.Fatal("expected pending to be invalid")
	}
}

func TestValidClassStatus(t *testing.T) {
	for _, s := range ClassStatuses {
		if !ValidClassStatus(string(s)) {
			t.Fatalf("ValidClassStatus(%q) = false, want true", s)
		}
	}
}

func TestValidSex(t *testing.T) {
	if !ValidSex("") || !ValidSex("M") || !ValidSex("F") {
		t.Fatal("expected empty/M/F to be valid")
	}
	if ValidSex("X") {
		t.Fatal("expected X to be invalid")
	}
}
