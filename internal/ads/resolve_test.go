package ads

import (
	"context"
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
	seen := make(map[int64]bool)
	for _, zone := range zones {
		name := cookieNameForAdZone(1, zone)
		id, ok := picks[name]
		if !ok {
			continue
		}
		if id != 10 && id != 20 && id != 30 {
			t.Fatalf("unexpected product id %d for zone %s", id, zone)
		}
		if seen[id] {
			t.Fatalf("duplicate product id %d across zones", id)
		}
		seen[id] = true
	}
	if len(picks) != 3 {
		t.Fatalf("pick count = %d, want 3 unique products for 4 zones", len(picks))
	}
}

func TestResolveCatalogDedupesProductsOnPage(t *testing.T) {
	t.Parallel()
	catalog := []CatalogAd{
		{
			ID:            1,
			Placement:     constants.AdPlacementAllSides,
			RandomizeKind: constants.AdRandomizePerPage,
			ProductOptions: []ProductOption{
				{ProductID: 10, Name: "A"},
				{ProductID: 20, Name: "B"},
			},
		},
		{
			ID:            2,
			Placement:     constants.AdPlacementTop,
			RandomizeKind: constants.AdRandomizePerPage,
			ProductOptions: []ProductOption{
				{ProductID: 10, Name: "A"},
				{ProductID: 30, Name: "C"},
			},
		},
	}
	r := httptest.NewRequest("GET", "/", nil)
	slots, _ := resolveCatalog(context.Background(), r, catalog, false)
	ids := collectResolvedProductIDs(slots)
	if len(ids) != 3 {
		t.Fatalf("expected 3 slot products, got %d (%v)", len(ids), ids)
	}
	if len(uniqueInts(ids)) != len(ids) {
		t.Fatalf("expected unique product ids on page, got %v", ids)
	}
}

func TestResolveCatalogForChromeRefreshKeepsPerPageStable(t *testing.T) {
	t.Parallel()
	catalog := []CatalogAd{
		{
			ID:            1,
			Placement:     constants.AdPlacementTop,
			RandomizeKind: constants.AdRandomizePerPage,
			ProductOptions: []ProductOption{
				{ProductID: 10},
				{ProductID: 20},
			},
		},
		{
			ID:            2,
			Placement:     constants.AdPlacementTop,
			RandomizeKind: constants.AdRandomizeTimer,
			TimerInterval: constants.AdTimerIntervalFiveSeconds,
			ProductOptions: []ProductOption{
				{ProductID: 30},
				{ProductID: 40},
			},
		},
	}
	page := httptest.NewRequest("GET", "/zion-english-admin/students", nil)
	page.Header.Set("Cookie", cookieNameForAdZonePerPage(1, constants.AdZoneTop, "/zion-english-admin/students")+"=10")
	first, _ := resolveCatalog(context.Background(), page, catalog, false)
	perPageID := first.Top[0].ProductID

	partial := httptest.NewRequest("GET", "/zion-english-admin/ads/partials/chrome", nil)
	partial.Header.Set("Referer", "http://localhost/zion-english-admin/students")
	partial.Header.Set("Cookie", page.Header.Get("Cookie"))
	second, _ := resolveCatalogForChromeRefresh(context.Background(), partial, catalog)
	if len(second.Top) != 2 {
		t.Fatalf("expected 2 top slots, got %d", len(second.Top))
	}
	var perPageAfter int64
	var timerAdID int64
	for _, p := range second.Top {
		if p.AdID == 1 {
			perPageAfter = p.ProductID
		}
		if p.AdID == 2 {
			timerAdID = p.ProductID
		}
	}
	if perPageAfter != perPageID {
		t.Fatalf("per-page product changed on chrome refresh: %d -> %d", perPageID, perPageAfter)
	}
	if timerAdID != 30 && timerAdID != 40 {
		t.Fatalf("unexpected timer product id %d", timerAdID)
	}
}

func TestResolveCatalogSkipsSlotsWhenPoolExhausted(t *testing.T) {
	t.Parallel()
	catalog := []CatalogAd{
		{
			ID:            1,
			Placement:     constants.AdPlacementAllSides,
			RandomizeKind: constants.AdRandomizePerPage,
			ProductOptions: []ProductOption{
				{ProductID: 10},
				{ProductID: 20},
			},
		},
	}
	r := httptest.NewRequest("GET", "/", nil)
	slots, _ := resolveCatalog(context.Background(), r, catalog, false)
	ids := collectResolvedProductIDs(slots)
	if len(ids) != 2 {
		t.Fatalf("expected 2 products with 2 options and 4 zones, got %d", len(ids))
	}
	if len(uniqueInts(ids)) != 2 {
		t.Fatalf("expected unique ids, got %v", ids)
	}
}

func collectResolvedProductIDs(slots ResolvedSlots) []int64 {
	var ids []int64
	for _, p := range slots.Top {
		ids = append(ids, p.ProductID)
	}
	for _, p := range slots.Left {
		ids = append(ids, p.ProductID)
	}
	for _, p := range slots.Right {
		ids = append(ids, p.ProductID)
	}
	for _, p := range slots.Bottom {
		ids = append(ids, p.ProductID)
	}
	return ids
}

func uniqueInts(vals []int64) []int64 {
	seen := make(map[int64]bool)
	var out []int64
	for _, v := range vals {
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}
