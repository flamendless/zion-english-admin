package frontend

import "zion-english/internal/constants"

type ParentPaymentVerificationStatus string

const (
	ParentPaymentVerified   ParentPaymentVerificationStatus = "verified"
	ParentPaymentMismatch   ParentPaymentVerificationStatus = "mismatch"
	ParentPaymentUnmatched  ParentPaymentVerificationStatus = "unmatched"
	ParentPaymentNoExpected ParentPaymentVerificationStatus = "no_expected"
)

type ParentPaymentVerificationRow struct {
	LegalName       string
	ParsedAmount    string
	StudentLabel    string
	ExpectedAmount  string
	Status          ParentPaymentVerificationStatus
	StatusLabel     string
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
	case ParentPaymentUnmatched, ParentPaymentNoExpected:
		return PillToneWarning
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
	case ParentPaymentUnmatched:
		return "No student"
	case ParentPaymentNoExpected:
		return "No classes"
	default:
		return constants.DisplayNotSet
	}
}
