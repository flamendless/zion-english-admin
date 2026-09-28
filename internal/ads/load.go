package ads

import (
	"context"
	"net/http"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
)

// LoadRequestContext loads published ads into ctx, applies rotation cookies, and resolves slots.
// When forceRotateFiveSecondTimers is true, timer ads with a 5 second interval get fresh picks.
func LoadRequestContext(ctx context.Context, w http.ResponseWriter, r *http.Request, db *queries.Queries, forceRotateFiveSecondTimers bool) context.Context {
	rows, err := db.GetPublishedAdsWithProducts(ctx)
	if err != nil {
		return ctx
	}
	catalog := buildCatalog(rows)
	if len(catalog) == 0 {
		return ctx
	}
	ctx = context.WithValue(ctx, catalogKey, catalog)
	slots, picks := resolveCatalog(ctx, r, catalog, forceRotateFiveSecondTimers)
	ctx = context.WithValue(ctx, rotationPicksKey, picks)
	ApplyRotationCookies(w, r, catalog, picks)
	return context.WithValue(ctx, slotsKey, slots)
}

// SuppressAdsForRequest reports whether ads should be hidden for the current viewer.
func SuppressAdsForRequest(ctx context.Context, r *http.Request, db *queries.Queries) bool {
	return suppressAdsForRequest(ctx, r, db)
}

// CatalogFastTimerRefreshSeconds returns the client poll interval when any ad uses a 5 second timer.
func CatalogFastTimerRefreshSeconds(catalog []CatalogAd) int {
	for _, ad := range catalog {
		if len(ad.ProductOptions) <= 1 {
			continue
		}
		if ad.RandomizeKind == constants.AdRandomizeTimer &&
			ad.TimerInterval == constants.AdTimerIntervalFiveSeconds {
			return 5
		}
	}
	return 0
}
