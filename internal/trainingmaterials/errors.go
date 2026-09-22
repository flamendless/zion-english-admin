package trainingmaterials

import "errors"

var (
	ErrTitleRequired       = errors.New("[TRAINING MATERIALS] title is required")
	ErrTitleTooLong        = errors.New("[TRAINING MATERIALS] title must be 64 characters or fewer")
	ErrDescriptionRequired = errors.New("[TRAINING MATERIALS] description is required")
	ErrURLRequired         = errors.New("[TRAINING MATERIALS] url is required")
	ErrInvalidURL          = errors.New("[TRAINING MATERIALS] url must be a valid YouTube link")
	ErrUnsupportedSourceType = errors.New("[TRAINING MATERIALS] unsupported source type")
	ErrInvalidStatus       = errors.New("[TRAINING MATERIALS] status must be published, draft, or deleted")
	ErrTagCount            = errors.New("[TRAINING MATERIALS] each material must have between 1 and 7 tags")
	ErrTagLabelTooLong     = errors.New("[TRAINING MATERIALS] tag labels must be 40 characters or fewer")
)
