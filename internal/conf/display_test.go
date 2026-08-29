package conf

import (
	"testing"

	"zion-english/internal/constants"
)

func TestMaskSecret(t *testing.T) {
	if got := MaskSecret(""); got != constants.StartupValueUnset {
		t.Fatalf("empty: got %q", got)
	}
	if got := MaskSecret("a"); got != "*" {
		t.Fatalf("len 1: got %q", got)
	}
	if got := MaskSecret("ab"); got != "**" {
		t.Fatalf("len 2: got %q", got)
	}
	if got := MaskSecret("abc"); got != "a**" {
		t.Fatalf("len 3: got %q", got)
	}
	if got := MaskSecret("abcd"); got != "a**d" {
		t.Fatalf("len 4: got %q", got)
	}
	if got := MaskSecret("abcde"); got != "ab**e" {
		t.Fatalf("len 5: got %q", got)
	}
	if got := MaskSecret("abcdefyz"); got != "ab*****z" {
		t.Fatalf("len 8: got %q", got)
	}
}
