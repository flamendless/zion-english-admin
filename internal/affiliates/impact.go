package affiliates

import (
	"fmt"
	"strings"

	"zion-english/internal/constants"
)

func ImpactDisplayAdThumbnailURL(programID, adID string) string {
	programID = strings.TrimSpace(programID)
	adID = strings.TrimSpace(adID)
	if programID == "" || adID == "" {
		return ""
	}
	return fmt.Sprintf("https://a.impactradius-go.com/display-ad/%s-%s", programID, adID)
}

type ProductThumbnailInput struct {
	Provider     constants.AffiliateProvider
	ItemID       string
	ProgramID    string
	ThumbnailURL string
}

func ResolveProductThumbnail(in ProductThumbnailInput) string {
	if in.Provider == constants.AffiliateProviderImpact {
		url := ImpactDisplayAdThumbnailURL(in.ProgramID, in.ItemID)
		if url != "" {
			return url
		}
	}
	return strings.TrimSpace(in.ThumbnailURL)
}
