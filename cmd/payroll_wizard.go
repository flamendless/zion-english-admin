package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/logs"
	"zion-english/internal/parentpaymentocr"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

type parentPaymentExpected struct {
	studentLabel string
	amount       float64
	currency     string
	hasExpected  bool
}

func handlePayrollWizardParentPaymentsOCR(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	writeHTML(w)

	if err := r.ParseMultipartForm(maxDocumentBytes); err != nil {
		renderParentPaymentsVerificationPanel(w, r, "", "", nil, true, "Upload failed or file is too large")
		return
	}

	startDate := strings.TrimSpace(r.FormValue("startDate"))
	endDate := strings.TrimSpace(r.FormValue("endDate"))
	if startDate == "" || endDate == "" {
		renderParentPaymentsVerificationPanel(w, r, startDate, endDate, nil, true, ErrMissingDateRange.Error())
		return
	}
	if startDate > endDate {
		renderParentPaymentsVerificationPanel(w, r, startDate, endDate, nil, true, ErrEndDateBeforeStart.Error())
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		renderParentPaymentsVerificationPanel(w, r, startDate, endDate, nil, true, "Please choose an image to upload")
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		renderParentPaymentsVerificationPanel(w, r, startDate, endDate, nil, true, "Failed to read uploaded image")
		return
	}

	reader := bytes.NewReader(imageBytes)
	ext, err := validateDocumentUpload(reader, header.Filename, int64(len(imageBytes)))
	if err != nil {
		renderParentPaymentsVerificationPanel(w, r, startDate, endDate, nil, true, userFacingUploadError(err))
		return
	}

	ctx := r.Context()
	text, ocrErr := parentpaymentocr.DetectDocumentText(ctx, imageBytes, ext)

	var entries []parentpaymentocr.PaymentOCREntry
	ocrFailed := false
	ocrMsg := ""
	if ocrErr != nil {
		logs.Log().Error("parent payments OCR", zap.Error(ocrErr))
		ocrFailed = true
		ocrMsg = ocrErr.Error()
		entries = []parentpaymentocr.PaymentOCREntry{{
			RawLine:    ocrMsg,
			ScanFailed: true,
		}}
	} else {
		entries = parentpaymentocr.ParsePaymentEntries(text)
		if len(entries) == 0 {
			entries = []parentpaymentocr.PaymentOCREntry{{
				RawLine: strings.TrimSpace(text),
				HasName: strings.TrimSpace(text) != "",
			}}
			if entries[0].HasName {
				entries[0].LegalName = entries[0].RawLine
			}
		}
	}

	expectedByLegal, err := loadExpectedParentPaymentsByLegalName(ctx, startDate, endDate)
	if err != nil {
		logs.Log().Error("load expected parent payments", zap.Error(err))
		renderParentPaymentsVerificationPanel(w, r, startDate, endDate, nil, true, "Failed to load expected parent totals")
		return
	}

	studentsByLegal, err := loadStudentLabelsByLegalName(ctx)
	if err != nil {
		logs.Log().Error("load student legal names", zap.Error(err))
		renderParentPaymentsVerificationPanel(w, r, startDate, endDate, nil, true, "Failed to load students")
		return
	}

	rows := buildParentPaymentVerificationRows(entries, expectedByLegal, studentsByLegal)
	renderParentPaymentsVerificationPanel(w, r, startDate, endDate, rows, ocrFailed, ocrMsg)
}

func renderParentPaymentsVerificationPanel(
	w http.ResponseWriter,
	r *http.Request,
	startDate, endDate string,
	rows []frontend.ParentPaymentVerificationRow,
	ocrFailed bool,
	ocrMsg string,
) {
	if len(rows) == 0 && ocrMsg != "" {
		rows = []frontend.ParentPaymentVerificationRow{{
			LegalName:      constants.DisplayNotSet,
			ParsedAmount:   constants.DisplayNotSet,
			StudentLabel:   constants.DisplayNotSet,
			ExpectedAmount: constants.DisplayNotSet,
			Status:         frontend.ParentPaymentOCRScanFailed,
			StatusLabel:    frontend.ParentPaymentStatusLabel(frontend.ParentPaymentOCRScanFailed),
		}}
		ocrFailed = true
	}

	cutoff := ""
	if startDate != "" && endDate != "" {
		cutoff = "Cutoff: " + formatReportCutoffLabel(startDate, endDate)
	}
	data := frontend.ParentPaymentsVerificationData{
		Rows:        rows,
		CutoffLabel: cutoff,
		OCRFailed:   ocrFailed,
		ErrorMsg:    ocrMsg,
	}
	if err := frontend.ParentPaymentsVerificationPanel(data).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func userFacingUploadError(err error) string {
	switch {
	case errors.Is(err, ErrInvalidDocumentImage), errors.Is(err, ErrUnsupportedDocumentFormat):
		return "Invalid image file. Please upload a PNG or JPEG image."
	case errors.Is(err, ErrDocumentFileTooLarge):
		return ErrDocumentFileTooLarge.Error()
	case errors.Is(err, ErrUploadedFileEmpty):
		return ErrUploadedFileEmpty.Error()
	default:
		return "Invalid upload"
	}
}

func loadStudentLabelsByLegalName(ctx context.Context) (map[string]string, error) {
	rows, err := dbRO.GetQueries().GetAllStudents(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, row := range rows {
		if !row.LegalName.Valid || strings.TrimSpace(row.LegalName.String) == "" {
			continue
		}
		key := parentpaymentocr.NormalizeLegalName(row.LegalName.String)
		if key == "" {
			continue
		}
		out[key] = utils.StudentDisplayNameFromNull(row.Name, row.StudentNumber, row.LegalName)
	}
	return out, nil
}

func loadExpectedParentPaymentsByLegalName(ctx context.Context, startDate, endDate string) (map[string]parentPaymentExpected, error) {
	rows, err := dbRO.GetQueries().GetReportSummaryRows(ctx, reportSummaryParams("", startDate, endDate, reportRoleFilters{}))
	if err != nil {
		return nil, err
	}

	type agg struct {
		studentName string
		studentNum  sql.NullString
		legalName   sql.NullString
		total       float64
		currency    string
		hasTotal    bool
	}
	byLegal := map[string]*agg{}

	for _, row := range rows {
		if !row.StudentLegalName.Valid || strings.TrimSpace(row.StudentLegalName.String) == "" {
			continue
		}
		key := parentpaymentocr.NormalizeLegalName(row.StudentLegalName.String)
		if key == "" {
			continue
		}
		classRow := summaryClassRowFromRecord(row)
		if !classRow.HasRate {
			continue
		}
		entry := byLegal[key]
		if entry == nil {
			entry = &agg{
				studentName: row.StudentName,
				studentNum:  row.StudentNumber,
				legalName:   row.StudentLegalName,
			}
			byLegal[key] = entry
		}
		entry.total += classRow.RateValue
		if row.ParentCurrency.Valid && row.ParentCurrency.String != "" {
			entry.currency = row.ParentCurrency.String
		}
		entry.hasTotal = true
	}

	out := make(map[string]parentPaymentExpected, len(byLegal))
	for key, entry := range byLegal {
		label := utils.StudentDisplayNameFromNull(entry.studentName, entry.studentNum, entry.legalName)
		out[key] = parentPaymentExpected{
			studentLabel: label,
			amount:       entry.total,
			currency:     entry.currency,
			hasExpected:  entry.hasTotal,
		}
	}
	return out, nil
}

func buildParentPaymentVerificationRows(
	entries []parentpaymentocr.PaymentOCREntry,
	expected map[string]parentPaymentExpected,
	studentsByLegal map[string]string,
) []frontend.ParentPaymentVerificationRow {
	rows := make([]frontend.ParentPaymentVerificationRow, 0, len(entries))
	for _, entry := range entries {
		row := frontend.ParentPaymentVerificationRow{
			ExpectedAmount: constants.DisplayNotSet,
			StudentLabel:   constants.DisplayNotSet,
		}

		if entry.ScanFailed {
			row.LegalName = constants.DisplayNotSet
			row.ParsedAmount = constants.DisplayNotSet
			row.Status = frontend.ParentPaymentOCRScanFailed
			row.StatusLabel = frontend.ParentPaymentStatusLabel(row.Status)
			rows = append(rows, row)
			continue
		}

		if entry.HasName {
			row.LegalName = entry.LegalName
		} else {
			row.LegalName = constants.DisplayNotSet
		}

		currency := ""
		if entry.HasName {
			key := parentpaymentocr.NormalizeLegalName(entry.LegalName)
			if exp, ok := expected[key]; ok && exp.currency != "" {
				currency = exp.currency
			}
		}

		if entry.HasAmount && entry.Amount > 0 {
			row.ParsedAmount = formatParsedAmount(entry.Amount, currency)
		} else {
			row.ParsedAmount = constants.DisplayNotSet
		}

		if !entry.HasName || !entry.HasAmount || entry.Amount <= 0 {
			if entry.HasAmount && entry.Amount > 0 && !entry.HasName {
				row.Status = frontend.ParentPaymentIncompleteOCR
			} else if entry.HasName && (!entry.HasAmount || entry.Amount <= 0) {
				row.Status = frontend.ParentPaymentIncompleteOCR
				key := parentpaymentocr.NormalizeLegalName(entry.LegalName)
				if label, ok := studentsByLegal[key]; ok {
					row.StudentLabel = label
				}
			} else {
				row.Status = frontend.ParentPaymentUnreadableOCR
				if row.LegalName == constants.DisplayNotSet && entry.RawLine != "" {
					row.LegalName = truncateOCRRaw(entry.RawLine, 80)
				}
			}
			row.StatusLabel = frontend.ParentPaymentStatusLabel(row.Status)
			rows = append(rows, row)
			continue
		}

		key := parentpaymentocr.NormalizeLegalName(entry.LegalName)
		exp, inCutoff := expected[key]
		studentLabel, inDB := studentsByLegal[key]

		if inDB {
			row.StudentLabel = studentLabel
		}

		if !inDB {
			row.Status = frontend.ParentPaymentNotInDatabase
			row.StatusLabel = frontend.ParentPaymentStatusLabel(row.Status)
			rows = append(rows, row)
			continue
		}

		if !inCutoff || !exp.hasExpected {
			row.Status = frontend.ParentPaymentNoExpected
			row.StatusLabel = frontend.ParentPaymentStatusLabel(row.Status)
			rows = append(rows, row)
			continue
		}

		row.ExpectedAmount = utils.FormatCurrencyAmountFixed(exp.amount, exp.currency) + " " + exp.currency
		if parentpaymentocr.AmountsMatch(entry.Amount, exp.amount) {
			row.Status = frontend.ParentPaymentVerified
		} else {
			row.Status = frontend.ParentPaymentMismatch
		}
		row.StatusLabel = frontend.ParentPaymentStatusLabel(row.Status)
		rows = append(rows, row)
	}
	return rows
}

func truncateOCRRaw(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func formatParsedAmount(amount float64, currency string) string {
	if currency != "" {
		return utils.FormatCurrencyAmountFixed(amount, currency) + " " + currency
	}
	return utils.FormatCurrencyAmountFixed(amount, "KRW")
}
