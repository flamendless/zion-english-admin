package announcements_test

import (
	"testing"

	"zion-english/internal/announcements"
)

func TestResolveCTAURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "external https",
			raw:  "https://example.com",
			want: "https://example.com",
		},
		{
			name: "external http",
			raw:  "http://example.com",
			want: "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := announcements.ResolveCTAURL(tt.raw)
			if got != tt.want {
				t.Fatalf("ResolveCTAURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestIsExternalCTAURL(t *testing.T) {
	if !announcements.IsExternalCTAURL("https://example.com") {
		t.Fatal("expected external https URL")
	}
	if announcements.IsExternalCTAURL("/students") {
		t.Fatal("expected internal path to be non-external")
	}
}
