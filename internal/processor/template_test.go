package processor

import "testing"

func TestValidateSheetTemplate(t *testing.T) {
	tests := []struct {
		in    string
		valid bool
	}{
		{"", true},
		{"A,B,C,G", true},
		{"A,C,D,H", true},
		{"A,B,C", false},
		{"A,B,C,G,H", false},
		{"1,B,C,G", false},
	}

	for _, tt := range tests {
		err := ValidateSheetTemplate(tt.in)
		if tt.valid && err != nil {
			t.Fatalf("ValidateSheetTemplate(%q) = %v, want nil", tt.in, err)
		}
		if !tt.valid && err == nil {
			t.Fatalf("ValidateSheetTemplate(%q) = nil, want error", tt.in)
		}
	}
}
