package learningmaterials

import "errors"

var (
	ErrTitleRequired       = errors.New("[LEARNING MATERIALS] title is required")
	ErrTitleTooLong        = errors.New("[LEARNING MATERIALS] title must be 64 characters or fewer")
	ErrDescriptionRequired = errors.New("[LEARNING MATERIALS] description is required")
	ErrURLRequired         = errors.New("[LEARNING MATERIALS] url is required")
	ErrInvalidURL          = errors.New("[LEARNING MATERIALS] url must be a valid http or https link")
	ErrInvalidAccess       = errors.New("[LEARNING MATERIALS] access must be public or private")
	ErrInvalidStatus       = errors.New("[LEARNING MATERIALS] status must be published, draft, or deleted")
	ErrTagCount            = errors.New("[LEARNING MATERIALS] each material must have between 1 and 7 tags")
	ErrTagLabelRequired    = errors.New("[LEARNING MATERIALS] tag labels cannot be empty")
	ErrTagLabelTooLong     = errors.New("[LEARNING MATERIALS] tag labels must be 40 characters or fewer")
)
