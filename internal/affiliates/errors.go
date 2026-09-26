package affiliates

import (
	"errors"
	"strings"
)

const affiliateErrPrefix = "[AFFILIATES] "

var (
	ErrAffiliateURLRequired = errors.New("[AFFILIATES] affiliate link is required")
	ErrInvalidAffiliateURL  = errors.New("[AFFILIATES] affiliate link must be a valid http or https URL")
	ErrNameRequired         = errors.New("[AFFILIATES] product name is required")
	ErrNameTooLong          = errors.New("[AFFILIATES] product name must be 256 characters or fewer")
	ErrInvalidThumbnailURL  = errors.New("[AFFILIATES] thumbnail url must be a valid http or https link")
	ErrInvalidProductURL    = errors.New("[AFFILIATES] product url must be a valid http or https link")
	ErrNotProductPage       = errors.New("[AFFILIATES] link does not resolve to a Shopee product page")
	ErrFetchBlocked         = errors.New("[AFFILIATES] could not load product page from Shopee")
	ErrPreviewUnavailable   = errors.New("[AFFILIATES] product preview is unavailable; fill in fields manually")
	ErrCSVInvalid           = errors.New("[AFFILIATES] invalid Shopee product links CSV")
	ErrCSVNoRows            = errors.New("[AFFILIATES] CSV has no product rows")
	ErrCSVTooManyRows       = errors.New("[AFFILIATES] CSV exceeds the maximum number of products per upload")
	ErrThumbnailNotFound    = errors.New("[AFFILIATES] thumbnail not found on product page")
)

func UserFacingMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if strings.HasPrefix(msg, affiliateErrPrefix) {
		return msg[len(affiliateErrPrefix):]
	}
	return msg
}
