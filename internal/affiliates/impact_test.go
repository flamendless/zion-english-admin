package affiliates

import (
	"testing"

	"zion-english/internal/constants"
)

func TestCatalogKeyImpactDoesNotCollideWithShopeeItemID(t *testing.T) {
	shopeeKey := CatalogKeyForProvider(constants.AffiliateProviderShopee, "3981287", "")
	impactKey := CatalogKeyForProvider(constants.AffiliateProviderImpact, "3981287", "")
	if shopeeKey == impactKey {
		t.Fatalf("keys collide: %s", shopeeKey)
	}
	if impactKey != "impact:3981287" {
		t.Fatalf("impact key: %s", impactKey)
	}
}

func TestImpactDisplayAdThumbnailURL(t *testing.T) {
	url := ImpactDisplayAdThumbnailURL("14579", "3981287")
	if url != "https://a.impactradius-go.com/display-ad/14579-3981287" {
		t.Fatalf("url: %s", url)
	}
}
