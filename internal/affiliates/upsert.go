package affiliates

import (
	"net/url"
	"strings"

	"zion-english/internal/constants"
)

func CatalogKey(itemID, productURL string) string {
	return CatalogKeyForProvider(constants.AffiliateProviderShopee, itemID, productURL)
}

func CatalogKeyForProvider(provider constants.AffiliateProvider, itemID, productURL string) string {
	if provider == constants.AffiliateProviderImpact {
		itemID = strings.TrimSpace(itemID)
		if itemID == "" {
			return ""
		}
		return "impact:" + itemID
	}
	itemID = strings.TrimSpace(itemID)
	if itemID != "" {
		return "item:" + itemID
	}
	norm := NormalizeProductURL(productURL)
	if norm == "" {
		return ""
	}
	return "url:" + norm
}

func NormalizeProductURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return strings.ToLower(raw)
	}
	parsed.Fragment = ""
	parsed.RawQuery = ""
	host := strings.ToLower(parsed.Host)
	path := strings.TrimSuffix(parsed.EscapedPath(), "/")
	if path == "" {
		path = "/"
	}
	return host + path
}

type CatalogRow struct {
	ID                   int64
	Provider             constants.AffiliateProvider
	ItemID               string
	ProductURL           string
	AffiliateURL         string
	Name                 string
	PriceDisplay         string
	Sales                string
	ShopBrandName        string
	CommissionRate       string
	Commission           string
	ThumbnailURL         string
	ProgramID            string
	ImpactState          string
	ImpactAdType         string
	ThumbnailOrientation string
}

func BuildCatalogIndex(rows []CatalogRow) map[string]CatalogRow {
	out := make(map[string]CatalogRow, len(rows))
	for _, row := range rows {
		key := CatalogKeyForProvider(row.Provider, row.ItemID, row.ProductURL)
		if key == "" {
			continue
		}
		if existing, ok := out[key]; ok && existing.ID < row.ID {
			continue
		}
		out[key] = row
	}
	return out
}

func FindExisting(index map[string]CatalogRow, provider constants.AffiliateProvider, itemID, productURL string) (CatalogRow, bool) {
	key := CatalogKeyForProvider(provider, itemID, productURL)
	if key == "" {
		return CatalogRow{}, false
	}
	row, ok := index[key]
	return row, ok
}
