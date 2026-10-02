package affiliates

import (
	"strings"
	"testing"

	"zion-english/internal/constants"
)

func TestOrientationFromCreativeDimensions(t *testing.T) {
	if OrientationFromCreativeDimensions("300", "200") != constants.ThumbnailOrientationLandscape {
		t.Fatal("expected landscape")
	}
	if OrientationFromCreativeDimensions("200", "300") != constants.ThumbnailOrientationPortrait {
		t.Fatal("expected portrait")
	}
	if OrientationFromCreativeDimensions("250", "250") != constants.ThumbnailOrientationSquare {
		t.Fatal("expected square")
	}
	if OrientationFromCreativeDimensions("", "100") != "" {
		t.Fatal("expected empty for missing width")
	}
	if OrientationFromCreativeDimensions("0", "100") != "" {
		t.Fatal("expected empty for zero width")
	}
}

func TestParseImpactAdsCSV_orientation(t *testing.T) {
	csv := `AdId,State,Name,ProgramId,AdType,TrackingLink,ThirdPartyServableAdCreativeWidth,ThirdPartyServableAdCreativeHeight
1,ACTIVE,Wide,1,IMAGE,https://example.com/a,320,200
2,ACTIVE,Tall,1,IMAGE,https://example.com/b,200,400
3,ACTIVE,Square,1,IMAGE,https://example.com/c,100,100
`
	rows, err := ParseImpactAdsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows: %d", len(rows))
	}
	if rows[0].Orientation != constants.ThumbnailOrientationLandscape {
		t.Fatalf("row0: %s", rows[0].Orientation)
	}
	if rows[1].Orientation != constants.ThumbnailOrientationPortrait {
		t.Fatalf("row1: %s", rows[1].Orientation)
	}
	if rows[2].Orientation != constants.ThumbnailOrientationSquare {
		t.Fatalf("row2: %s", rows[2].Orientation)
	}
}
