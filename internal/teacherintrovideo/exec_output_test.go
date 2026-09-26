package teacherintrovideo

import (
	"strings"
	"testing"
)

func TestTrimExecOutput(t *testing.T) {
	cases := []struct {
		name string
		out  []byte
		want string
	}{
		{"empty", nil, "(no output)"},
		{"whitespace", []byte("  \n\t "), "(no output)"},
		{"single line", []byte("Error: invalid data\n"), "Error: invalid data"},
		{"collapses fields", []byte("line one\nline two"), "line one line two"},
	}
	for _, tc := range cases {
		if got := trimExecOutput(tc.out); got != tc.want {
			t.Fatalf("%s: trimExecOutput() = %q, want %q", tc.name, got, tc.want)
		}
	}
	long := strings.Repeat("x", maxExecOutputLogChars+50)
	got := trimExecOutput([]byte(long))
	if len(got) != maxExecOutputLogChars+3 || !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncated output with suffix, got len %d", len(got))
	}
}
