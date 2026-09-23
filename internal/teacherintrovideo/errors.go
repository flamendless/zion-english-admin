package teacherintrovideo

import "errors"

var (
	ErrURLRequired           = errors.New("[INTRO VIDEO] url is required")
	ErrInvalidURL            = errors.New("[INTRO VIDEO] url must be a valid link for the selected source")
	ErrUnsupportedSourceType = errors.New("[INTRO VIDEO] source type must be google drive or youtube")
)
