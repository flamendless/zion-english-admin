package cmd

import (
	"context"
	"testing"

	"zion-english/internal/parentpaymentocr"
)

func TestScanParentPaymentImageWithoutTesseract(t *testing.T) {
	if parentpaymentocr.TesseractInstalled() {
		t.Skip("tesseract installed; skipping missing-binary case")
	}
	res := scanParentPaymentImage(context.Background(), []byte{0x89, 0x50, 0x4e, 0x47}, ".png")
	if !res.OCRFailed {
		t.Fatal("expected OCRFailed when tesseract is missing")
	}
	if res.OCRMsg == "" {
		t.Fatal("expected OCR error message")
	}
	if len(res.Entries) != 1 || !res.Entries[0].ScanFailed {
		t.Fatal("expected single scan-failed entry marker")
	}
}
