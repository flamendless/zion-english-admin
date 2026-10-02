package affiliates

import (
	"strings"
	"testing"
)

func TestParseImpactAdsCSV_minimal(t *testing.T) {
	csv := `AdId,State,Name,ProgramId,AdType,TrackingLink
3981287,ACTIVE,Test Ad,14579,IMAGE,https://englishonline.sjv.io/c/7848748/3981287/14579
`
	rows, err := ParseImpactAdsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows: %d", len(rows))
	}
	if rows[0].AdID != "3981287" || rows[0].ProgramID != "14579" {
		t.Fatalf("row: %+v", rows[0])
	}
}

func TestParseImpactAdsCSV_missingColumn(t *testing.T) {
	csv := `AdId,State,Name,ProgramId,AdType
1,ACTIVE,X,1,IMAGE
`
	_, err := ParseImpactAdsCSV(strings.NewReader(csv))
	if err != ErrCSVInvalid {
		t.Fatalf("expected ErrCSVInvalid, got %v", err)
	}
}
