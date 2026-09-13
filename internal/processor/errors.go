package processor

import "errors"

var (
	ErrInvalidSheetTemplate = errors.New("[PROCESSOR] template must be four comma-separated Excel column letters (e.g. A,B,C,G)")
	ErrEmptyRecord          = errors.New("[PROCESSOR] empty record")
	ErrInvalidTime          = errors.New("[PROCESSOR] invalid time")
)
