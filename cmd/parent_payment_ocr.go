package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/logs"
	"zion-english/internal/parentpaymentocr"

	"go.uber.org/zap"
)

type parentPaymentScanResult struct {
	RawText   string
	Entries   []parentpaymentocr.PaymentOCREntry
	OCRFailed bool
	OCRMsg    string
}

func scanParentPaymentImage(ctx context.Context, imageBytes []byte, ext string) parentPaymentScanResult {
	text, ocrErr := parentpaymentocr.DetectDocumentText(ctx, imageBytes, ext)
	if ocrErr != nil {
		logs.Log().Error("parent payments OCR", zap.Error(ocrErr))
		return parentPaymentScanResult{
			Entries: []parentpaymentocr.PaymentOCREntry{{
				RawLine:    ocrErr.Error(),
				ScanFailed: true,
			}},
			OCRFailed: true,
			OCRMsg:    ocrErr.Error(),
		}
	}

	entries := parentpaymentocr.ParsePaymentEntries(text)
	if len(entries) == 0 {
		entries = []parentpaymentocr.PaymentOCREntry{{
			RawLine: strings.TrimSpace(text),
			HasName: strings.TrimSpace(text) != "",
		}}
		if entries[0].HasName {
			entries[0].LegalName = entries[0].RawLine
		}
	}
	return parentPaymentScanResult{
		RawText: text,
		Entries: entries,
	}
}

func handleParentPaymentOCRTool(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	data := frontend.ParentPaymentOCRToolData{
		TesseractInstalled: parentpaymentocr.TesseractInstalled(),
		TesseractLanguages: parentpaymentocr.ResolvedTesseractLanguages(),
	}
	if err := frontend.ParentPaymentOCRTool(data).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleParentPaymentOCRToolScan(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	writeHTML(w)

	if err := r.ParseMultipartForm(maxDocumentBytes); err != nil {
		renderParentPaymentOCRToolScan(w, r, parentPaymentScanResult{
			OCRFailed: true,
			OCRMsg:    "Upload failed or file is too large",
		})
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		renderParentPaymentOCRToolScan(w, r, parentPaymentScanResult{
			OCRFailed: true,
			OCRMsg:    "Please choose an image to upload",
		})
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		renderParentPaymentOCRToolScan(w, r, parentPaymentScanResult{
			OCRFailed: true,
			OCRMsg:    "Failed to read uploaded image",
		})
		return
	}

	reader := bytes.NewReader(imageBytes)
	ext, err := validateDocumentUpload(reader, header.Filename, int64(len(imageBytes)))
	if err != nil {
		renderParentPaymentOCRToolScan(w, r, parentPaymentScanResult{
			OCRFailed: true,
			OCRMsg:    userFacingUploadError(err),
		})
		return
	}

	scan := scanParentPaymentImage(r.Context(), imageBytes, ext)
	renderParentPaymentOCRToolScan(w, r, scan)
}

func displayOCROtherDetails(details string) string {
	details = strings.TrimSpace(details)
	if details == "" {
		return constants.DisplayNotSet
	}
	return details
}

func renderParentPaymentOCRToolScan(w http.ResponseWriter, r *http.Request, scan parentPaymentScanResult) {
	rows := make([]frontend.ParentPaymentOCRToolEntryRow, 0, len(scan.Entries))
	for _, entry := range scan.Entries {
		row := frontend.ParentPaymentOCRToolEntryRow{
			RawLine: entry.RawLine,
		}
		if entry.ScanFailed {
			row.LegalName = constants.DisplayNotSet
			row.Amount = constants.DisplayNotSet
			row.CompleteLabel = constants.DisplayNotSet
			rows = append(rows, row)
			continue
		}
		if entry.HasName && strings.TrimSpace(entry.LegalName) != "" {
			row.LegalName = entry.LegalName
		} else {
			row.LegalName = constants.DisplayNotSet
		}
		if entry.HasAmount && entry.Amount > 0 {
			row.Amount = formatParsedAmount(entry.Amount, "")
		} else {
			row.Amount = constants.DisplayNotSet
		}
		if entry.Complete() {
			row.CompleteLabel = "Yes"
		} else {
			row.CompleteLabel = "No"
		}
		row.OtherDetails = displayOCROtherDetails(entry.OtherDetails)
		rows = append(rows, row)
	}

	data := frontend.ParentPaymentOCRToolScanData{
		OCRFailed: scan.OCRFailed,
		ErrorMsg:  scan.OCRMsg,
		RawText:   scan.RawText,
		Rows:      rows,
	}
	if err := frontend.ParentPaymentOCRToolScanResults(data).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}
