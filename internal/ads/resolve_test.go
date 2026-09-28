package ads

import (
	"net/http/httptest"
	"testing"

	"zion-english/internal/constants"
)

func TestCookieNameForAdZone(t *testing.T) {
	got := cookieNameForAdZone(42, constants.AdZoneLeft)
	want := "zion_ad_42_left"
	if got != want {
		t.Fatalf("cookie name = %q, want %q", got, want)
	}
}

func TestComputeRotationPicksPerZone(t *testing.T) {
	catalog := []CatalogAd{
		{
			ID:            1,
			Placement:     constants.AdPlacementAllSides,
			RandomizeKind: constants.AdRandomizePerSession,
			ProductOptions: []ProductOption{
				{ProductID: 10},
				{ProductID: 20},
				{ProductID: 30},
			},
		},
	}
	r := httptest.NewRequest("GET", "/", nil)
	picks := ComputeRotationPicks(r, catalog)
	zones := []constants.AdZone{
		constants.AdZoneTop,
		constants.AdZoneLeft,
		constants.AdZoneRight,
		constants.AdZoneBottom,
	}
	for _, zone := range zones {
		name := cookieNameForAdZone(1, zone)
		id, ok := picks[name]
		if !ok {
			t.Fatalf("missing pick for zone %s", zone)
		}
		if id != 10 && id != 20 && id != 30 {
			t.Fatalf("unexpected product id %d for zone %s", id, zone)
		}
	}
	if len(picks) != 4 {
		t.Fatalf("pick count = %d, want 4", len(picks))
	}
}
