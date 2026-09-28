package ads

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"zion-english/internal/conf"
	"zion-english/internal/constants"
)

const cookiePrefix = "zion_ad_"

var resolutionZoneOrder = []constants.AdZone{
	constants.AdZoneTop,
	constants.AdZoneLeft,
	constants.AdZoneRight,
	constants.AdZoneBottom,
}

type ResolvedProduct struct {
	AdID         int64
	AdName       string
	Zone         constants.AdZone
	ProductID    int64
	Name         string
	ShopName     string
	PriceDisplay string
	ThumbnailURL string
	AffiliateURL string
	Sales        string
}

type ResolvedSlots struct {
	Top          []ResolvedProduct
	Left         []ResolvedProduct
	Right        []ResolvedProduct
	Bottom       []ResolvedProduct
	HasSideRails bool
}

func ResolveSlots(ctx context.Context, r *http.Request) ResolvedSlots {
	catalog := GetCatalog(ctx)
	if len(catalog) == 0 {
		return ResolvedSlots{}
	}
	slots, _ := resolveCatalog(ctx, r, catalog, false)
	return slots
}

func resolveCatalog(ctx context.Context, r *http.Request, catalog []CatalogAd, forceRotateFiveSecondTimers bool) (ResolvedSlots, map[string]int64) {
	used := make(map[int64]bool)
	picks := make(map[string]int64)
	var slots ResolvedSlots

	for _, ad := range catalog {
		for _, zone := range zonesForAdInResolutionOrder(ad) {
			opt, ok := pickProductForSlot(ctx, r, ad, zone, used, picks, forceRotateFiveSecondTimers)
			if !ok {
				continue
			}
			used[opt.ProductID] = true
			if len(ad.ProductOptions) > 1 {
				switch ad.RandomizeKind {
				case constants.AdRandomizePerSession, constants.AdRandomizeTimer:
					picks[cookieNameForAdZone(ad.ID, zone)] = opt.ProductID
				}
			}
			product := resolvedProductFromOption(ad, zone, opt)
			appendResolvedProduct(&slots, product)
		}
	}
	return slots, picks
}

func zonesForAdInResolutionOrder(ad CatalogAd) []constants.AdZone {
	inPlacement := make(map[constants.AdZone]bool)
	for _, z := range ad.Placement.Zones() {
		inPlacement[z] = true
	}
	out := make([]constants.AdZone, 0, len(inPlacement))
	for _, z := range resolutionZoneOrder {
		if inPlacement[z] {
			out = append(out, z)
		}
	}
	return out
}

func appendResolvedProduct(slots *ResolvedSlots, product ResolvedProduct) {
	switch product.Zone {
	case constants.AdZoneTop:
		slots.Top = append(slots.Top, product)
	case constants.AdZoneLeft:
		slots.Left = append(slots.Left, product)
		slots.HasSideRails = true
	case constants.AdZoneRight:
		slots.Right = append(slots.Right, product)
		slots.HasSideRails = true
	case constants.AdZoneBottom:
		slots.Bottom = append(slots.Bottom, product)
	}
}

func resolvedProductFromOption(ad CatalogAd, zone constants.AdZone, chosen ProductOption) ResolvedProduct {
	return ResolvedProduct{
		AdID:         ad.ID,
		AdName:       ad.Name,
		Zone:         zone,
		ProductID:    chosen.ProductID,
		Name:         chosen.Name,
		ShopName:     chosen.ShopName,
		PriceDisplay: chosen.PriceDisplay,
		ThumbnailURL: chosen.ThumbnailURL,
		AffiliateURL: chosen.AffiliateURL,
		Sales:        chosen.Sales,
	}
}

func pickProductForSlot(
	ctx context.Context,
	r *http.Request,
	ad CatalogAd,
	zone constants.AdZone,
	used map[int64]bool,
	picks map[string]int64,
	forceRotateFiveSecondTimers bool,
) (ProductOption, bool) {
	unused := filterUnusedOptions(ad.ProductOptions, used)
	if len(unused) == 0 {
		return ProductOption{}, false
	}
	if len(ad.ProductOptions) == 1 {
		return unused[0], true
	}

	switch ad.RandomizeKind {
	case constants.AdRandomizePerPage:
		return unused[randomIndex(len(unused))], true
	case constants.AdRandomizePerSession, constants.AdRandomizeTimer:
		cookieName := cookieNameForAdZone(ad.ID, zone)
		if forceRotateFiveSecondTimers &&
			ad.RandomizeKind == constants.AdRandomizeTimer &&
			ad.TimerInterval == constants.AdTimerIntervalFiveSeconds {
			return unused[randomIndex(len(unused))], true
		}
		if id, ok := stableProductIDForSlot(ctx, r, cookieName, picks); ok {
			if opt, found := findProduct(unused, id); found {
				return opt, true
			}
		}
		return unused[randomIndex(len(unused))], true
	default:
		return unused[randomIndex(len(unused))], true
	}
}

func stableProductIDForSlot(ctx context.Context, r *http.Request, cookieName string, picks map[string]int64) (int64, bool) {
	if c, err := r.Cookie(cookieName); err == nil {
		if id, err := strconv.ParseInt(c.Value, 10, 64); err == nil {
			return id, true
		}
	}
	if id, ok := picks[cookieName]; ok {
		return id, true
	}
	if id, ok := rotationPickFromContext(ctx, cookieName); ok {
		return id, true
	}
	return 0, false
}

func filterUnusedOptions(options []ProductOption, used map[int64]bool) []ProductOption {
	if len(used) == 0 {
		return options
	}
	out := make([]ProductOption, 0, len(options))
	for _, opt := range options {
		if !used[opt.ProductID] {
			out = append(out, opt)
		}
	}
	return out
}

func findProduct(options []ProductOption, productID int64) (ProductOption, bool) {
	for _, opt := range options {
		if opt.ProductID == productID {
			return opt, true
		}
	}
	return ProductOption{}, false
}

func cookieNameForAdZone(adID int64, zone constants.AdZone) string {
	return fmt.Sprintf("%s%d_%s", cookiePrefix, adID, zone)
}

func timerMaxAge(interval constants.AdTimerInterval) int {
	switch interval {
	case constants.AdTimerIntervalFiveSeconds:
		return 5
	case constants.AdTimerIntervalHourly:
		return 3600
	case constants.AdTimerIntervalDaily:
		return 86400
	default:
		return 3600
	}
}

func randomIndex(n int) int {
	if n <= 1 {
		return 0
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return int(time.Now().UnixNano()) % n
	}
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n))
}

// ApplyRotationCookies sets HttpOnly cookies for session/timer rotation picks when needed.
func ApplyRotationCookies(w http.ResponseWriter, r *http.Request, catalog []CatalogAd, picks map[string]int64) {
	cookiePath := conf.Conf().BasePath
	if cookiePath == "" {
		cookiePath = "/"
	}
	for _, ad := range catalog {
		if len(ad.ProductOptions) <= 1 {
			continue
		}
		for _, zone := range zonesForAdInResolutionOrder(ad) {
			cookieName := cookieNameForAdZone(ad.ID, zone)
			switch ad.RandomizeKind {
			case constants.AdRandomizePerSession:
				if _, err := r.Cookie(cookieName); err != nil {
					productID, ok := picks[cookieName]
					if !ok {
						continue
					}
					http.SetCookie(w, &http.Cookie{
						Name:     cookieName,
						Value:    strconv.FormatInt(productID, 10),
						Path:     cookiePath,
						HttpOnly: true,
						SameSite: http.SameSiteLaxMode,
					})
				}
			case constants.AdRandomizeTimer:
				if _, err := r.Cookie(cookieName); err != nil {
					productID, ok := picks[cookieName]
					if !ok {
						continue
					}
					http.SetCookie(w, &http.Cookie{
						Name:     cookieName,
						Value:    strconv.FormatInt(productID, 10),
						Path:     cookiePath,
						MaxAge:   timerMaxAge(ad.TimerInterval),
						HttpOnly: true,
						SameSite: http.SameSiteLaxMode,
					})
				}
			}
		}
	}
}

// ComputeRotationPicks chooses stable session/timer products per ad zone for this request.
func ComputeRotationPicks(r *http.Request, catalog []CatalogAd) map[string]int64 {
	_, picks := resolveCatalog(context.Background(), r, catalog, false)
	return picks
}

func rotationPickFromContext(ctx context.Context, cookieName string) (int64, bool) {
	picks, _ := ctx.Value(rotationPicksKey).(map[string]int64)
	if picks == nil {
		return 0, false
	}
	id, ok := picks[cookieName]
	return id, ok
}
