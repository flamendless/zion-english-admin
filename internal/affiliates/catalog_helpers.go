package affiliates

import (
	"zion-english/internal/constants"
)

func CatalogRowFromDB(
	id int64,
	provider string,
	itemID, productURL, affiliateURL, name string,
	priceDisplay, sales, shopBrandName, commissionRate, commission, thumbnailURL string,
	programID, impactState, impactAdType, thumbnailOrientation string,
) CatalogRow {
	p := constants.AffiliateProvider(provider)
	if !constants.ValidAffiliateProvider(provider) {
		p = constants.AffiliateProviderShopee
	}
	return CatalogRow{
		ID:                   id,
		Provider:             p,
		ItemID:               itemID,
		ProductURL:           productURL,
		AffiliateURL:         affiliateURL,
		Name:                 name,
		PriceDisplay:         priceDisplay,
		Sales:                sales,
		ShopBrandName:        shopBrandName,
		CommissionRate:       commissionRate,
		Commission:           commission,
		ThumbnailURL:         thumbnailURL,
		ProgramID:            programID,
		ImpactState:          impactState,
		ImpactAdType:         impactAdType,
		ThumbnailOrientation: thumbnailOrientation,
	}
}
