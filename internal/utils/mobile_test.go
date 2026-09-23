package utils

import "testing"

func TestValidMobileNumber(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"09171234567", true},
		{"+639171234567", true},
		{"+63 917 123 4567", true},
		{"(0917) 123-4567", true},
		{"917-123-4567", true},
		{"<img src=x onerror='alert(document.domain)'>", false},
		{"abc", false},
		{"123", false},
		{"", false},
		{"1234567890123456", false},
	}
	for _, tc := range tests {
		got := ValidMobileNumber(tc.in)
		if got != tc.want {
			t.Fatalf("ValidMobileNumber(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
