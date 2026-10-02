package affiliates

import (
	"strings"
	"testing"
)

func TestValidImpactCreativePart(t *testing.T) {
	if !ValidImpactCreativePart("14579") {
		t.Fatal("expected numeric program id valid")
	}
	if ValidImpactCreativePart("") || ValidImpactCreativePart("12a") || ValidImpactCreativePart("12-3") {
		t.Fatal("expected invalid ids rejected")
	}
}

func TestImpactCreativeImagePath(t *testing.T) {
	if got := ImpactCreativeImagePath(42, "", ""); got != "/affiliate-creative/42" {
		t.Fatalf("product path: %q", got)
	}
	got := ImpactCreativeImagePath(0, "14579", "3981287")
	if !strings.Contains(got, "program_id=14579") || !strings.Contains(got, "ad_id=3981287") {
		t.Fatalf("query path: %q", got)
	}
}

func TestResolveProductThumbnailImpactUsesProxyPath(t *testing.T) {
	path := ImpactCreativeImagePath(9, "14579", "3981287")
	if path != "/affiliate-creative/9" {
		t.Fatalf("expected product proxy path, got %q", path)
	}
}
