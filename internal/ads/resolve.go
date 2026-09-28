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

type ResolvedProduct struct {
	AdID          int64
	AdName        string
	ProductID     int64
	Name          string
	ShopName      string
	PriceDisplay  string
	ThumbnailURL  string
	AffiliateURL  string
	Sales         string
}

type ResolvedSlots struct {
	Top    []ResolvedProduct
	Left   []ResolvedProduct
	Right  []ResolvedProduct
	Bottom []ResolvedProduct
	HasSideRails bool
}

func ResolveSlots(ctx context.Context, r *http.Request) ResolvedSlots {
	catalog := GetCatalog(ctx)
	if len(catalog) == 0 {
		return ResolvedSlots{}
	}

	var slots ResolvedSlots
	for _, ad := range catalog {
		for _, zone := range ad.Placement.Zones() {
			product := resolveProductForAdZone(ctx, ad, zone, r)
			if product == nil {
				continue
			}
			switch zone {
			case constants.AdZoneTop:
				slots.Top = append(slots.Top, *product)
			case constants.AdZoneLeft:
				slots.Left = append(slots.Left, *product)
				slots.HasSideRails = true
			case constants.AdZoneRight:
				slots.Right = append(slots.Right, *product)
				slots.HasSideRails = true
			case constants.AdZoneBottom:
				slots.Bottom = append(slots.Bottom, *product)
			}
		}
	}
	return slots
}

func resolveProductForAdZone(ctx context.Context, ad CatalogAd, zone constants.AdZone, r *http.Request) *ResolvedProduct {
	if len(ad.ProductOptions) == 0 {
		return nil
	}
	var chosen ProductOption
	if len(ad.ProductOptions) == 1 {
		chosen = ad.ProductOptions[0]
	} else {
		switch ad.RandomizeKind {
		case constants.AdRandomizePerPage:
			chosen = ad.ProductOptions[randomIndex(len(ad.ProductOptions))]
		case constants.AdRandomizePerSession:
			chosen = resolveSessionProduct(ctx, ad, zone, r)
		case constants.AdRandomizeTimer:
			chosen = resolveTimerProduct(ctx, ad, zone, r)
		default:
			chosen = ad.ProductOptions[randomIndex(len(ad.ProductOptions))]
		}
	}
	return &ResolvedProduct{
		AdID:         ad.ID,
		AdName:       ad.Name,
		ProductID:    chosen.ProductID,
		Name:         chosen.Name,
		ShopName:     chosen.ShopName,
		PriceDisplay: chosen.PriceDisplay,
		ThumbnailURL: chosen.ThumbnailURL,
		AffiliateURL: chosen.AffiliateURL,
		Sales:        chosen.Sales,
	}
}

func resolveSessionProduct(ctx context.Context, ad CatalogAd, zone constants.AdZone, r *http.Request) ProductOption {
	cookieName := cookieNameForAdZone(ad.ID, zone)
	if c, err := r.Cookie(cookieName); err == nil {
		if id, err := strconv.ParseInt(c.Value, 10, 64); err == nil {
			if opt, ok := findProduct(ad.ProductOptions, id); ok {
				return opt
			}
		}
	}
	if id, ok := rotationPickFromContext(ctx, cookieName); ok {
		if opt, found := findProduct(ad.ProductOptions, id); found {
			return opt
		}
	}
	return ad.ProductOptions[randomIndex(len(ad.ProductOptions))]
}

func resolveTimerProduct(ctx context.Context, ad CatalogAd, zone constants.AdZone, r *http.Request) ProductOption {
	cookieName := cookieNameForAdZone(ad.ID, zone)
	if c, err := r.Cookie(cookieName); err == nil {
		if id, err := strconv.ParseInt(c.Value, 10, 64); err == nil {
			if opt, ok := findProduct(ad.ProductOptions, id); ok {
				return opt
			}
		}
	}
	if id, ok := rotationPickFromContext(ctx, cookieName); ok {
		if opt, found := findProduct(ad.ProductOptions, id); found {
			return opt
		}
	}
	return ad.ProductOptions[randomIndex(len(ad.ProductOptions))]
}

func rotationPickFromContext(ctx context.Context, cookieName string) (int64, bool) {
	picks, _ := ctx.Value(rotationPicksKey).(map[string]int64)
	if picks == nil {
		return 0, false
	}
	id, ok := picks[cookieName]
	return id, ok
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
		for _, zone := range ad.Placement.Zones() {
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
	picks := make(map[string]int64)
	for _, ad := range catalog {
		if len(ad.ProductOptions) <= 1 {
			continue
		}
		switch ad.RandomizeKind {
		case constants.AdRandomizePerSession, constants.AdRandomizeTimer:
		default:
			continue
		}
		for _, zone := range ad.Placement.Zones() {
			cookieName := cookieNameForAdZone(ad.ID, zone)
			if c, err := r.Cookie(cookieName); err == nil {
				if id, err := strconv.ParseInt(c.Value, 10, 64); err == nil {
					if _, ok := findProduct(ad.ProductOptions, id); ok {
						picks[cookieName] = id
						continue
					}
				}
			}
			chosen := ad.ProductOptions[randomIndex(len(ad.ProductOptions))]
			picks[cookieName] = chosen.ProductID
		}
	}
	return picks
}
