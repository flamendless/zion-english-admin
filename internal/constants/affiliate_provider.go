package constants

type AffiliateProvider string

const (
	AffiliateProviderShopee AffiliateProvider = "shopee"
	AffiliateProviderImpact AffiliateProvider = "impact"
)

func ValidAffiliateProvider(value string) bool {
	switch AffiliateProvider(value) {
	case AffiliateProviderShopee, AffiliateProviderImpact:
		return true
	default:
		return false
	}
}

func AffiliateProviderLabel(provider AffiliateProvider) string {
	switch provider {
	case AffiliateProviderShopee:
		return "Shopee"
	case AffiliateProviderImpact:
		return "Impact"
	default:
		return string(provider)
	}
}
