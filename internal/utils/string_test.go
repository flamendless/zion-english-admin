package utils

import "testing"

func TestSanitizeSheetName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Don Teacher", "Don Teacher"},
		{"Name/With\\Bad?Chars*[1]", "NameWithBadChars1"},
		{"", ""},
	}
	for _, tc := range tests {
		got := SanitizeSheetName(tc.in)
		if got != tc.want {
			t.Fatalf("SanitizeSheetName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	long := "abcdefghijklmnopqrstuvwxyz1234567890"
	got := SanitizeSheetName(long)
	if len(got) != 31 {
		t.Fatalf("SanitizeSheetName long name len = %d, want 31", len(got))
	}
}

func TestTeacherCompactDisplayName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"-", "-"},
		{"Maria", "Maria"},
		{"John Doe", "John D"},
		{"Maria Ana Santos", "Maria S"},
		{"  Juan  Dela  Cruz  ", "Juan C"},
	}
	for _, tc := range tests {
		got := TeacherCompactDisplayName(tc.in)
		if got != tc.want {
			t.Fatalf("TeacherCompactDisplayName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
