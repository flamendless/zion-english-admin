package affiliates

import (
	"fmt"
	"strconv"
	"strings"

	"zion-english/internal/constants"
)

func OrientationFromCreativeDimensions(widthRaw, heightRaw string) constants.ThumbnailOrientation {
	w, errW := strconv.Atoi(strings.TrimSpace(widthRaw))
	h, errH := strconv.Atoi(strings.TrimSpace(heightRaw))
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return ""
	}
	if w > h {
		return constants.ThumbnailOrientationLandscape
	}
	if h > w {
		return constants.ThumbnailOrientationPortrait
	}
	return constants.ThumbnailOrientationSquare
}

func ImpactDisplayAdThumbnailURL(programID, adID string) string {
	programID = strings.TrimSpace(programID)
	adID = strings.TrimSpace(adID)
	if programID == "" || adID == "" {
		return ""
	}
	return fmt.Sprintf("https://a.impactradius-go.com/display-ad/%s-%s", programID, adID)
}

type ProductThumbnailInput struct {
	ProductID    int64
	Provider     constants.AffiliateProvider
	ItemID       string
	ProgramID    string
	ThumbnailURL string
}

func ResolveProductThumbnail(in ProductThumbnailInput) string {
	if in.Provider == constants.AffiliateProviderImpact {
		url := ImpactCreativeImageURL(in.ProductID, in.ProgramID, in.ItemID)
		if url != "" {
			return url
		}
	}
	return strings.TrimSpace(in.ThumbnailURL)
}
