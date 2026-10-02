package affiliates

import (
	"testing"

	"zion-english/internal/constants"
)

func TestCatalogKey(t *testing.T) {
	if CatalogKey("123", "") != "item:123" {
		t.Fatal("expected item key")
	}
	if CatalogKey("", "https://shopee.ph/product/1") == "" {
		t.Fatal("expected url key")
	}
}

func TestDiffImportRowNewAndChanged(t *testing.T) {
	index := BuildCatalogIndex([]CatalogRow{
		{
			ID:           1,
			Provider:     constants.AffiliateProviderShopee,
			ItemID:       "99",
			ProductURL:   "https://shopee.ph/p/99",
			AffiliateURL: "https://offer/1",
			Name:         "Old name",
			PriceDisplay: "100",
		},
	})
	seen := make(map[string]bool)

	newRes := DiffImportRow(ImportRowInput{
		Provider:     constants.AffiliateProviderShopee,
		ItemID:       "100",
		ProductURL:   "https://shopee.ph/p/100",
		OfferLink:    "https://offer/2",
		ItemName:     "New",
		PriceDisplay: "50",
	}, index, seen)
	if newRes.Status != constants.AffiliateImportDiffNew || !newRes.IncludeByDefault {
		t.Fatalf("new: %+v", newRes)
	}

	changedRes := DiffImportRow(ImportRowInput{
		Provider:     constants.AffiliateProviderShopee,
		ItemID:       "99",
		ProductURL:   "https://shopee.ph/p/99",
		OfferLink:    "https://offer/1",
		ItemName:     "New name",
		PriceDisplay: "100",
	}, index, seen)
	if changedRes.Status != constants.AffiliateImportDiffChanged || changedRes.ExistingID != 1 {
		t.Fatalf("changed: %+v", changedRes)
	}
	if len(changedRes.ChangedFields) == 0 {
		t.Fatal("expected changed fields")
	}
}

func TestDiffImportRowDuplicateInCSV(t *testing.T) {
	index := map[string]CatalogRow{}
	seen := make(map[string]bool)
	row := ImportRowInput{Provider: constants.AffiliateProviderShopee, ItemID: "1", ProductURL: "https://a", OfferLink: "https://b", ItemName: "X", PriceDisplay: "1"}
	DiffImportRow(row, index, seen)
	dup := DiffImportRow(row, index, seen)
	if dup.Status != constants.AffiliateImportDiffDuplicateInCSV {
		t.Fatalf("dup: %+v", dup)
	}
}

func TestRemovedFromCSV(t *testing.T) {
	catalog := []CatalogRow{
		{ID: 1, Provider: constants.AffiliateProviderShopee, ItemID: "a", Name: "A"},
		{ID: 2, Provider: constants.AffiliateProviderShopee, ItemID: "b", Name: "B"},
		{ID: 3, Provider: constants.AffiliateProviderShopee, ItemID: "", Name: "No ID"},
	}
	removed := RemovedFromCSV(catalog, map[string]bool{"a": true}, constants.AffiliateProviderShopee)
	if len(removed) != 1 || removed[0].ItemID != "b" {
		t.Fatalf("removed: %+v", removed)
	}
}
