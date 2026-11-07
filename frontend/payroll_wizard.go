package frontend

import "zion-english/internal/constants"

type ParentPaymentVerificationStatus string

const (
	ParentPaymentVerified      ParentPaymentVerificationStatus = "verified"
	ParentPaymentMismatch      ParentPaymentVerificationStatus = "mismatch"
	ParentPaymentNotInDatabase ParentPaymentVerificationStatus = "not_in_database"
	ParentPaymentNoExpected    ParentPaymentVerificationStatus = "no_expected"
	ParentPaymentIncompleteOCR ParentPaymentVerificationStatus = "incomplete_ocr"
	ParentPaymentUnreadableOCR ParentPaymentVerificationStatus = "unreadable_ocr"
	ParentPaymentOCRScanFailed ParentPaymentVerificationStatus = "scan_failed"
)

type ParentPaymentVerificationRow struct {
	LegalName      string
	ParsedAmount   string
	StudentLabel   string
	ExpectedAmount string
	Status         ParentPaymentVerificationStatus
	StatusLabel    string
}

type ParentPaymentsVerificationData struct {
	Rows        []ParentPaymentVerificationRow
	CutoffLabel string
	OCRFailed   bool
	ErrorMsg    string
}

func (s ParentPaymentVerificationStatus) PillTone() PillTone {
	switch s {
	case ParentPaymentVerified:
		return PillToneSuccess
	case ParentPaymentMismatch:
		return PillToneError
	case ParentPaymentNotInDatabase, ParentPaymentNoExpected, ParentPaymentIncompleteOCR:
		return PillToneWarning
	case ParentPaymentUnreadableOCR, ParentPaymentOCRScanFailed:
		return PillToneNeutral
	default:
		return PillToneNeutral
	}
}

func ParentPaymentStatusLabel(status ParentPaymentVerificationStatus) string {
	switch status {
	case ParentPaymentVerified:
		return "Match"
	case ParentPaymentMismatch:
		return "Mismatch"
	case ParentPaymentNotInDatabase:
		return "Not in database"
	case ParentPaymentNoExpected:
		return "No classes"
	case ParentPaymentIncompleteOCR:
		return "Incomplete OCR"
	case ParentPaymentUnreadableOCR:
		return "Unreadable"
	case ParentPaymentOCRScanFailed:
		return "Scan failed"
	default:
		return constants.DisplayNotSet
	}
}
