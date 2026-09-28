package ads

import (
	"context"
	"net/http"
	"strings"

	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
)

type contextKey string

const catalogKey contextKey = "ads_catalog"
const slotsKey contextKey = "ads_resolved_slots"
const rotationPicksKey contextKey = "ads_rotation_picks"

type CatalogAd struct {
	ID             int64
	Name           string
	Placement      constants.AdPlacement
	AdType         constants.AdType
	RandomizeKind  constants.AdRandomizeKind
	TimerInterval  constants.AdTimerInterval
	SortOrder      int64
	ProductOptions []ProductOption
}

type ProductOption struct {
	ProductID     int64
	Name          string
	ShopName      string
	PriceDisplay  string
	ThumbnailURL  string
	AffiliateURL  string
	Sales         string
	SortOrder     int64
}

func GetCatalog(ctx context.Context) []CatalogAd {
	catalog, _ := ctx.Value(catalogKey).([]CatalogAd)
	return catalog
}

func GetResolvedSlots(ctx context.Context) ResolvedSlots {
	slots, _ := ctx.Value(slotsKey).(ResolvedSlots)
	return slots
}

func Middleware(db *queries.Queries, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !shouldLoadAds(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		rows, err := db.GetPublishedAdsWithProducts(ctx)
		if err == nil {
			catalog := buildCatalog(rows)
			ctx = context.WithValue(ctx, catalogKey, catalog)
			picks := ComputeRotationPicks(r, catalog)
			ctx = context.WithValue(ctx, rotationPicksKey, picks)
			ApplyRotationCookies(w, r, catalog, picks)
			slots := ResolveSlots(ctx, r)
			ctx = context.WithValue(ctx, slotsKey, slots)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func shouldLoadAds(path string) bool {
	if isLandingPath(path) {
		return false
	}
	if strings.Contains(path, "/static/") {
		return false
	}
	if strings.HasSuffix(path, "/health") {
		return false
	}
	if strings.Contains(path, "/api/") {
		return false
	}
	if strings.Contains(path, "/partials/") {
		return false
	}
	if strings.Contains(path, "/unread-count") {
		return false
	}
	if strings.Contains(path, "/notifications/panel") {
		return false
	}
	return true
}

func isLandingPath(path string) bool {
	base := strings.TrimSuffix(conf.Conf().BasePath, "/")
	p := strings.TrimSuffix(path, "/")
	if base == "" {
		return p == "" || p == "/"
	}
	return p == base
}

func buildCatalog(rows []queries.GetPublishedAdsWithProductsRow) []CatalogAd {
	if len(rows) == 0 {
		return nil
	}
	byID := make(map[int64]*CatalogAd)
	order := make([]int64, 0)
	for _, row := range rows {
		ad, ok := byID[row.AdID]
		if !ok {
			ad = &CatalogAd{
				ID:            row.AdID,
				Name:          row.AdName,
				Placement:     constants.AdPlacement(row.Placement),
				AdType:        constants.AdType(row.AdType),
				RandomizeKind: constants.AdRandomizeKind(row.RandomizeKind),
				TimerInterval: constants.AdTimerInterval(row.TimerInterval),
				SortOrder:     row.AdSortOrder,
			}
			byID[row.AdID] = ad
			order = append(order, row.AdID)
		}
		ad.ProductOptions = append(ad.ProductOptions, ProductOption{
			ProductID:    row.AffiliateProductID,
			Name:         row.ProductName,
			ShopName:     row.ShopName,
			Sales:        row.Sales,
			PriceDisplay: row.PriceDisplay,
			ThumbnailURL: row.ThumbnailUrl,
			AffiliateURL: row.AffiliateUrl,
			SortOrder:    row.ProductSortOrder,
		})
	}
	out := make([]CatalogAd, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out
}
