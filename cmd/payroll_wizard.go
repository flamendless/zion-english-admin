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
		renderParentPaymentsVerificationError(w, r, "Upload failed or file is too large")
		return
	}

	startDate := strings.TrimSpace(r.FormValue("startDate"))
	endDate := strings.TrimSpace(r.FormValue("endDate"))
	if startDate == "" || endDate == "" {
		renderParentPaymentsVerificationError(w, r, ErrMissingDateRange.Error())
		return
	}
	if startDate > endDate {
		renderParentPaymentsVerificationError(w, r, ErrEndDateBeforeStart.Error())
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		renderParentPaymentsVerificationError(w, r, "Please choose an image to upload")
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		renderParentPaymentsVerificationError(w, r, "Failed to read uploaded image")
		return
	}

	reader := bytes.NewReader(imageBytes)
	ext, err := validateDocumentUpload(reader, header.Filename, int64(len(imageBytes)))
	if err != nil {
		renderParentPaymentsVerificationError(w, r, userFacingUploadError(err))
		return
	}

	ctx := r.Context()
	text, ocrErr := parentpaymentocr.DetectDocumentText(ctx, imageBytes, ext)
	if ocrErr != nil {
		logs.Log().Error("parent payments OCR", zap.Error(ocrErr))
		renderParentPaymentsVerificationError(w, r, ocrErr.Error())
		return
	}

	parsed := parentpaymentocr.ParsePaymentLines(text)
	expectedByLegal, err := loadExpectedParentPaymentsByLegalName(ctx, startDate, endDate)
	if err != nil {
		logs.Log().Error("load expected parent payments", zap.Error(err))
		renderParentPaymentsVerificationError(w, r, "Failed to load expected parent totals")
		return
	}

	rows := buildParentPaymentVerificationRows(parsed, expectedByLegal)
	data := frontend.ParentPaymentsVerificationData{
		Rows:        rows,
		CutoffLabel: "Cutoff: " + formatReportCutoffLabel(startDate, endDate),
	}
	if err := frontend.ParentPaymentsVerificationPanel(data).Render(ctx, w); err != nil {
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

func renderParentPaymentsVerificationError(w http.ResponseWriter, r *http.Request, message string) {
	data := frontend.ParentPaymentsVerificationData{
		OCRFailed: true,
		ErrorMsg:  message,
	}
	if err := frontend.ParentPaymentsVerificationPanel(data).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
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

func buildParentPaymentVerificationRows(parsed []parentpaymentocr.ParsedLine, expected map[string]parentPaymentExpected) []frontend.ParentPaymentVerificationRow {
	rows := make([]frontend.ParentPaymentVerificationRow, 0, len(parsed))
	for _, line := range parsed {
		key := parentpaymentocr.NormalizeLegalName(line.LegalName)
		exp, ok := expected[key]
		row := frontend.ParentPaymentVerificationRow{
			LegalName:    line.LegalName,
			ParsedAmount: formatParsedAmount(line.Amount, exp.currency),
		}
		if !ok || !exp.hasExpected {
			row.Status = frontend.ParentPaymentUnmatched
			if ok && !exp.hasExpected {
				row.Status = frontend.ParentPaymentNoExpected
				row.StudentLabel = exp.studentLabel
			}
			row.ExpectedAmount = constants.DisplayNotSet
		} else {
			row.StudentLabel = exp.studentLabel
			row.ExpectedAmount = utils.FormatCurrencyAmountFixed(exp.amount, exp.currency) + " " + exp.currency
			if parentpaymentocr.AmountsMatch(line.Amount, exp.amount) {
				row.Status = frontend.ParentPaymentVerified
			} else {
				row.Status = frontend.ParentPaymentMismatch
			}
		}
		row.StatusLabel = frontend.ParentPaymentStatusLabel(row.Status)
		rows = append(rows, row)
	}
	return rows
}

func formatParsedAmount(amount float64, currency string) string {
	if currency != "" {
		return utils.FormatCurrencyAmountFixed(amount, currency) + " " + currency
	}
	return utils.FormatCurrencyAmountFixed(amount, "KRW")
}
