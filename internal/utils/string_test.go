package utils

import "testing"

func TestComposePersonName(t *testing.T) {
	tests := []struct {
		first  string
		middle string
		last   string
		want   string
	}{
		{"Jane", "", "Doe", "Jane Doe"},
		{"Jane", "Marie", "Doe", "Jane Marie Doe"},
		{" Jane ", " Marie ", " Doe ", "Jane Marie Doe"},
		{"Jane", "", "", "Jane"},
	}

	for _, tt := range tests {
		got := ComposePersonName(tt.first, tt.middle, tt.last)
		if got != tt.want {
			t.Fatalf("ComposePersonName(%q, %q, %q) = %q, want %q", tt.first, tt.middle, tt.last, got, tt.want)
		}
	}
}
