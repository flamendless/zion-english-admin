package utils

import "testing"

func TestProfileNameEditable(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"Jane", false},
		{"   ", false},
		{"-", false},
	}

	for _, tt := range tests {
		got := ProfileNameEditable(tt.in)
		if got != tt.want {
			t.Fatalf("ProfileNameEditable(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestIsBlank(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"\t", true},
		{"Jane", false},
		{" Jane ", false},
	}

	for _, tt := range tests {
		got := IsBlank(tt.in)
		if got != tt.want {
			t.Fatalf("IsBlank(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"  Foo@Bar.COM  ", "foo@bar.com"},
		{"teacher@example.com", "teacher@example.com"},
	}

	for _, tt := range tests {
		got := NormalizeEmail(tt.in)
		if got != tt.want {
			t.Fatalf("NormalizeEmail(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

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
