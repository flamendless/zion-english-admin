package affiliates

import (
	"net/url"
	"strings"
)

func CatalogKey(itemID, productURL string) string {
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
	ID             int64
	ItemID         string
	ProductURL     string
	AffiliateURL   string
	Name           string
	PriceDisplay   string
	Sales          string
	ShopBrandName  string
	CommissionRate string
	Commission     string
	ThumbnailURL   string
}

func BuildCatalogIndex(rows []CatalogRow) map[string]CatalogRow {
	out := make(map[string]CatalogRow, len(rows))
	for _, row := range rows {
		key := CatalogKey(row.ItemID, row.ProductURL)
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

func FindExisting(index map[string]CatalogRow, itemID, productURL string) (CatalogRow, bool) {
	key := CatalogKey(itemID, productURL)
	if key == "" {
		return CatalogRow{}, false
	}
	row, ok := index[key]
	return row, ok
}
