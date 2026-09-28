package constants

type AdType string

const (
	AdTypeAffiliate AdType = "affiliate"
)

func ValidAdType(adType string) bool {
	switch AdType(adType) {
	case AdTypeAffiliate:
		return true
	default:
		return false
	}
}
