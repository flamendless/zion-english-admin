package metatags

import "errors"

var (
	ErrNameRequired           = errors.New("[META] name is required")
	ErrContentOrValueRequired = errors.New("[META] content or value is required")
	ErrInvalidName            = errors.New("[META] invalid meta name")
	ErrAttributeTooLong       = errors.New("[META] content or value is too long")
	ErrDuplicateName          = errors.New("[META] a tag with this name already exists")
)
