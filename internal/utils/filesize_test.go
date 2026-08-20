package utils

import "testing"

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{5 << 20, "5.0 MB"},
	}
	for _, tt := range tests {
		if got := FormatFileSize(tt.bytes); got != tt.want {
			t.Fatalf("FormatFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}
