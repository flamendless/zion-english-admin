package affiliates

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
)

type ClickAttribution struct {
	AdID int64
	Zone constants.AdZone
}

type ClickAttributionQuerier interface {
	AdAffiliateProductInPublishedAd(ctx context.Context, arg queries.AdAffiliateProductInPublishedAdParams) (int64, error)
	GetAdByID(ctx context.Context, id int64) (queries.TblAd, error)
}

func ResolveClickAttribution(ctx context.Context, q ClickAttributionQuerier, productID int64, adIDRaw, zoneRaw string) (ClickAttribution, bool) {
	adIDStr := strings.TrimSpace(adIDRaw)
	zoneStr := strings.TrimSpace(zoneRaw)
	if adIDStr == "" || zoneStr == "" {
		return ClickAttribution{}, false
	}
	if !constants.ValidAdZone(zoneStr) {
		return ClickAttribution{}, false
	}
	adID, err := strconv.ParseInt(adIDStr, 10, 64)
	if err != nil || adID <= 0 || productID <= 0 {
		return ClickAttribution{}, false
	}
	zone := constants.AdZone(zoneStr)

	_, err = q.AdAffiliateProductInPublishedAd(ctx, queries.AdAffiliateProductInPublishedAdParams{
		ID:                 adID,
		AffiliateProductID: productID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ClickAttribution{}, false
		}
		return ClickAttribution{}, false
	}

	ad, err := q.GetAdByID(ctx, adID)
	if err != nil {
		return ClickAttribution{}, false
	}
	for _, z := range constants.AdPlacement(ad.Placement).Zones() {
		if z == zone {
			return ClickAttribution{AdID: adID, Zone: zone}, true
		}
	}
	return ClickAttribution{}, false
}
