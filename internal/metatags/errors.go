package metatags

import "errors"

var (
	ErrNameRequired           = errors.New("[META] name is required")
	ErrContentOrValueRequired = errors.New("[META] content or value is required")
	ErrInvalidName            = errors.New("[META] invalid meta name")
	ErrAttributeTooLong       = errors.New("[META] content or value is too long")
	ErrDuplicateName          = errors.New("[META] a tag with this name already exists")
	ErrReservedName           = errors.New("[META] this name is reserved; use each page title instead of a meta tag")
	ErrInvalidAttr            = errors.New("[META] invalid attribute type")
	ErrInvalidScope           = errors.New("[META] invalid render scope")
	ErrInvalidOgImageURL      = errors.New("[META] og:image content must be a valid http or https URL")
)
