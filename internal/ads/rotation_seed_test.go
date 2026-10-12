package ads

import (
	"testing"

	"zion-english/internal/constants"
)

func TestDeterministicProductStableForSlot(t *testing.T) {
	t.Parallel()
	unused := []ProductOption{{ProductID: 10}, {ProductID: 20}, {ProductID: 30}}
	ad := CatalogAd{ID: 1, RandomizeKind: constants.AdRandomizePerSession}
	slotKey := rotationSlotKey(ad, constants.AdZoneTop, nil)
	a, ok := deterministicProductInUnused("seed-a", slotKey, unused)
	if !ok {
		t.Fatal("expected pick")
	}
	b, ok := deterministicProductInUnused("seed-a", slotKey, unused)
	if !ok || a != b {
		t.Fatalf("expected stable pick, got %d and %d", a, b)
	}
	c, ok := deterministicProductInUnused("seed-b", slotKey, unused)
	if !ok {
		t.Fatal("expected pick for other seed")
	}
	if a == c && len(unused) > 1 {
		t.Fatalf("expected different seeds to differ when possible, both %d", a)
	}
}
