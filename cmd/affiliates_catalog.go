package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"zion-english/internal/affiliates"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
)

func loadAffiliateCatalog(ctx context.Context) ([]affiliates.CatalogRow, map[string]affiliates.CatalogRow, error) {
	rows, err := dbRO.GetQueries().GetAffiliateProductsCatalogKeys(ctx)
	if err != nil {
		return nil, nil, err
	}
	catalog := make([]affiliates.CatalogRow, 0, len(rows))
	for _, row := range rows {
		catalog = append(catalog, affiliates.CatalogRow{
			ID:             row.ID,
			ItemID:         row.ItemID,
			ProductURL:     row.ProductUrl,
			AffiliateURL:   row.AffiliateUrl,
			Name:           row.Name,
			PriceDisplay:   row.PriceDisplay,
			Sales:          row.Sales,
			ShopBrandName:  row.ShopBrandName,
			CommissionRate: row.CommissionRate,
			Commission:     row.Commission,
			ThumbnailURL:   row.ThumbnailUrl,
		})
	}
	return catalog, affiliates.BuildCatalogIndex(catalog), nil
}

func resolveAffiliateUpsertTargetID(index map[string]affiliates.CatalogRow, itemID, productURL string, currentID int64) int64 {
	existing, ok := affiliates.FindExisting(index, itemID, productURL)
	if !ok {
		return 0
	}
	if currentID > 0 && currentID == existing.ID {
		return currentID
	}
	return existing.ID
}

type affiliateProductWriteParams struct {
	AffiliateURL     string
	ProductURL       string
	ShopID           string
	ItemID           string
	Name             string
	Brand            string
	PriceDisplay     string
	ThumbnailURL     string
	SortOrder        int64
	ImportBatchID    int64
	Sales            string
	AffiliatedShopID sql.NullInt64
	CommissionRate   string
	Commission       string
}

func upsertAffiliateProduct(ctx context.Context, db database.Service, index map[string]affiliates.CatalogRow, currentID int64, p affiliateProductWriteParams) (id int64, isUpdate bool, merged bool, err error) {
	targetID := resolveAffiliateUpsertTargetID(index, p.ItemID, p.ProductURL, currentID)
	if targetID == 0 && currentID > 0 {
		targetID = currentID
	}
	q := db.GetQueries()
	if targetID > 0 {
		err = q.UpdateAffiliateProduct(ctx, queries.UpdateAffiliateProductParams{
			AffiliateUrl:     p.AffiliateURL,
			ProductUrl:       p.ProductURL,
			ShopID:           p.ShopID,
			ItemID:           p.ItemID,
			Name:             p.Name,
			Brand:            p.Brand,
			PriceDisplay:     p.PriceDisplay,
			ThumbnailUrl:     p.ThumbnailURL,
			SortOrder:        p.SortOrder,
			Sales:            p.Sales,
			AffiliatedShopID: p.AffiliatedShopID,
			CommissionRate:   p.CommissionRate,
			Commission:       p.Commission,
			ID:               targetID,
		})
		if err != nil {
			return 0, false, false, err
		}
		merged = currentID > 0 && targetID != currentID
		return targetID, true, merged, nil
	}
	id, err = q.InsertAffiliateProduct(ctx, queries.InsertAffiliateProductParams{
		AffiliateUrl:     p.AffiliateURL,
		ProductUrl:       p.ProductURL,
		ShopID:           p.ShopID,
		ItemID:           p.ItemID,
		Name:             p.Name,
		Brand:            p.Brand,
		PriceDisplay:     p.PriceDisplay,
		ThumbnailUrl:     p.ThumbnailURL,
		SortOrder:        p.SortOrder,
		ImportBatchID:    p.ImportBatchID,
		Sales:            p.Sales,
		AffiliatedShopID: p.AffiliatedShopID,
		CommissionRate:   p.CommissionRate,
		Commission:       p.Commission,
	})
	if err != nil {
		return 0, false, false, err
	}
	return id, false, false, nil
}

func parseStagedExistingID(r *http.Request, index int) int64 {
	prefix := fmt.Sprintf("staged_%d_", index)
	raw := strings.TrimSpace(r.FormValue(prefix + "existing_id"))
	if raw == "" {
		return 0
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}
