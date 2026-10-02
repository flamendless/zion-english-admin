package affiliates

import (
	"strings"

	"zion-english/internal/constants"
)

type ImportRowInput struct {
	Provider       constants.AffiliateProvider
	ItemID         string
	ProductURL     string
	OfferLink      string
	ItemName       string
	PriceDisplay   string
	Sales          string
	ShopName       string
	CommissionRate string
	Commission     string
	ThumbnailURL   string
	ProgramID      string
	ImpactState    string
	ImpactAdType   string
}

type ImportDiffResult struct {
	Status           constants.AffiliateImportDiffStatus
	ExistingID       int64
	ChangedFields    []string
	IncludeByDefault bool
}

func DiffImportRow(row ImportRowInput, index map[string]CatalogRow, seenKeys map[string]bool) ImportDiffResult {
	if row.Provider == "" {
		row.Provider = constants.AffiliateProviderShopee
	}
	key := CatalogKeyForProvider(row.Provider, row.ItemID, row.ProductURL)
	if key == "" {
		return ImportDiffResult{
			Status:           constants.AffiliateImportDiffMissingItemKey,
			IncludeByDefault: false,
		}
	}
	if seenKeys[key] {
		return ImportDiffResult{
			Status:           constants.AffiliateImportDiffDuplicateInCSV,
			IncludeByDefault: false,
		}
	}
	seenKeys[key] = true

	existing, ok := index[key]
	if !ok {
		return ImportDiffResult{
			Status:           constants.AffiliateImportDiffNew,
			IncludeByDefault: true,
		}
	}

	changed := compareImportFields(row, existing)
	if len(changed) == 0 {
		return ImportDiffResult{
			Status:           constants.AffiliateImportDiffUnchanged,
			ExistingID:       existing.ID,
			IncludeByDefault: false,
		}
	}
	return ImportDiffResult{
		Status:           constants.AffiliateImportDiffChanged,
		ExistingID:       existing.ID,
		ChangedFields:    changed,
		IncludeByDefault: true,
	}
}

func compareImportFields(row ImportRowInput, existing CatalogRow) []string {
	var changed []string
	if !strings.EqualFold(strings.TrimSpace(row.ItemName), strings.TrimSpace(existing.Name)) {
		changed = append(changed, "Name")
	}
	if strings.TrimSpace(row.OfferLink) != strings.TrimSpace(existing.AffiliateURL) {
		changed = append(changed, "Tracking link")
	}
	if row.Provider == constants.AffiliateProviderImpact {
		if !strings.EqualFold(strings.TrimSpace(row.ImpactState), strings.TrimSpace(existing.ImpactState)) {
			changed = append(changed, "State")
		}
		if strings.TrimSpace(row.ProgramID) != strings.TrimSpace(existing.ProgramID) {
			changed = append(changed, "Program ID")
		}
		if !strings.EqualFold(strings.TrimSpace(row.ImpactAdType), strings.TrimSpace(existing.ImpactAdType)) {
			changed = append(changed, "Ad type")
		}
		return changed
	}
	if NormalizeProductURL(row.ProductURL) != NormalizeProductURL(existing.ProductURL) {
		changed = append(changed, "Product link")
	}
	if strings.TrimSpace(row.PriceDisplay) != strings.TrimSpace(existing.PriceDisplay) {
		changed = append(changed, "Price")
	}
	if strings.TrimSpace(row.Sales) != strings.TrimSpace(existing.Sales) {
		changed = append(changed, "Sales")
	}
	if strings.TrimSpace(row.ShopName) != strings.TrimSpace(existing.ShopBrandName) {
		changed = append(changed, "Shop")
	}
	if strings.TrimSpace(row.CommissionRate) != strings.TrimSpace(existing.CommissionRate) {
		changed = append(changed, "Commission rate")
	}
	if strings.TrimSpace(row.Commission) != strings.TrimSpace(existing.Commission) {
		changed = append(changed, "Commission")
	}
	return changed
}

type RemovedCatalogItem struct {
	ID     int64
	ItemID string
	Name   string
}

func RemovedFromCSV(catalog []CatalogRow, csvItemIDs map[string]bool, provider constants.AffiliateProvider) []RemovedCatalogItem {
	out := make([]RemovedCatalogItem, 0)
	for _, row := range catalog {
		if row.Provider != provider {
			continue
		}
		itemID := strings.TrimSpace(row.ItemID)
		if itemID == "" {
			continue
		}
		if csvItemIDs[itemID] {
			continue
		}
		out = append(out, RemovedCatalogItem{
			ID:     row.ID,
			ItemID: itemID,
			Name:   row.Name,
		})
	}
	return out
}

func CollectCSVItemIDs(rows []ImportRowInput) map[string]bool {
	out := make(map[string]bool)
	for _, row := range rows {
		itemID := strings.TrimSpace(row.ItemID)
		if itemID != "" {
			out[itemID] = true
		}
	}
	return out
}
