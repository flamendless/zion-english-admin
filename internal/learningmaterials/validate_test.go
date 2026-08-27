package learningmaterials

import "testing"

func TestNormalizeTagLabels(t *testing.T) {
	got := NormalizeTagLabels([]string{" Grammar ", "grammar", "VOCAB", "", "reading"})
	want := []string{"grammar", "vocab", "reading"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestValidateRequest(t *testing.T) {
	err := ValidateRequest(Request{
		Description: "A useful worksheet",
		URL:         "https://example.com/resource",
		Access:      AccessPublic,
		Status:      StatusDraft,
		TagLabels:   []string{"grammar"},
	})
	if err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}

	err = ValidateRequest(Request{
		Description: "Missing tags",
		URL:         "https://example.com/resource",
		Access:      AccessPublic,
		Status:      StatusDraft,
		TagLabels:   nil,
	})
	if err != ErrTagCount {
		t.Fatalf("expected ErrTagCount, got %v", err)
	}
}
