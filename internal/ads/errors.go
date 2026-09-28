package ads

import "errors"

var (
	ErrNameRequired              = errors.New("[ADS] name is required")
	ErrInvalidPlacement          = errors.New("[ADS] invalid placement")
	ErrInvalidAdType             = errors.New("[ADS] invalid ad type")
	ErrInvalidStatus             = errors.New("[ADS] status must be published, draft, or deleted")
	ErrInvalidFormStatus         = errors.New("[ADS] status must be published or draft")
	ErrInvalidRandomizeKind      = errors.New("[ADS] invalid randomize kind")
	ErrInvalidTimerInterval      = errors.New("[ADS] timer interval must be 5 seconds, hourly, or daily when randomize kind is timer")
	ErrTimerIntervalNotAllowed   = errors.New("[ADS] timer interval is only allowed when randomize kind is timer")
	ErrAffiliateProductsRequired = errors.New("[ADS] at least one affiliate product is required for affiliate ads")
	ErrTooManyProductsForKind    = errors.New("[ADS] randomize kind is only used when multiple affiliate products are selected")
	ErrRandomizeKindRequired     = errors.New("[ADS] randomize kind is required when multiple affiliate products are selected")
)
