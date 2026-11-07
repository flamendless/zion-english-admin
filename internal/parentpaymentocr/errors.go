package parentpaymentocr

import "errors"

var (
	ErrVisionAPIKeyMissing = errors.New("[PARENT PAYMENTS] Google Vision API key is not configured")
	ErrVisionRequestFailed = errors.New("[PARENT PAYMENTS] OCR request failed")
	ErrVisionNoText        = errors.New("[PARENT PAYMENTS] No text found in image")
)
