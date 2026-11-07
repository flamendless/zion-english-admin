package parentpaymentocr

import "errors"

var (
	ErrTesseractNotInstalled = errors.New("[PARENT PAYMENTS] Tesseract OCR is not installed (install tesseract-ocr and language packs)")
	ErrOCRFailed             = errors.New("[PARENT PAYMENTS] OCR failed")
	ErrOCRNoText             = errors.New("[PARENT PAYMENTS] No text found in image")
)
