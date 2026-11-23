package frontend

type ParentPaymentOCRToolData struct {
	TesseractInstalled bool
	TesseractLanguages string
}

type ParentPaymentOCRToolEntryRow struct {
	LegalName     string
	Amount        string
	CompleteLabel string
	RawLine       string
}

type ParentPaymentOCRToolScanData struct {
	OCRFailed bool
	ErrorMsg  string
	RawText   string
	Rows      []ParentPaymentOCRToolEntryRow
}