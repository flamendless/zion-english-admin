package ads

import (
	"context"
	"net/http"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
)

// LoadRequestContext loads published ads into ctx, applies rotation cookies, and resolves slots.
func LoadRequestContext(ctx context.Context, w http.ResponseWriter, r *http.Request, db *queries.Queries) context.Context {
	rows, err := db.GetPublishedAdsWithProducts(ctx)
	if err != nil {
		return ctx
	}
	catalog := buildCatalog(rows)
	if len(catalog) == 0 {
		return ctx
	}
	ctx = context.WithValue(ctx, catalogKey, catalog)
	slots, picks := resolveCatalog(ctx, r, catalog, false)
	ctx = context.WithValue(ctx, rotationPicksKey, picks)
	ApplyRotationCookies(w, r, catalog, picks, false)
	return context.WithValue(ctx, slotsKey, slots)
}

// LoadChromePartialRequestContext resolves slots for the fast chrome refresh: only 5 second timer ads rotate.
func LoadChromePartialRequestContext(ctx context.Context, w http.ResponseWriter, r *http.Request, db *queries.Queries) context.Context {
	rows, err := db.GetPublishedAdsWithProducts(ctx)
	if err != nil {
		return ctx
	}
	catalog := buildCatalog(rows)
	if len(catalog) == 0 {
		return ctx
	}
	ctx = context.WithValue(ctx, catalogKey, catalog)
	slots, picks := resolveCatalogForChromeRefresh(ctx, r, catalog)
	ctx = context.WithValue(ctx, rotationPicksKey, picks)
	ApplyRotationCookies(w, r, catalog, picks, true)
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
