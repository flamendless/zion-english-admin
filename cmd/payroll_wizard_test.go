package cmd

import (
	"testing"

	"zion-english/frontend"
	"zion-english/internal/parentpaymentocr"
)

func TestBuildParentPaymentVerificationRowsNotInDatabase(t *testing.T) {
	entries := []parentpaymentocr.PaymentOCREntry{{
		LegalName: "없는학생",
		Amount:    50000,
		HasName:   true,
		HasAmount: true,
		RawLine:   "test",
	}}
	expected := map[string]parentPaymentExpected{}
	students := map[string]string{"김현주": "Student A (김현주)"}

	rows := buildParentPaymentVerificationRows(entries, expected, students)
	if len(rows) != 1 {
		t.Fatalf("rows: %+v", rows)
	}
	if rows[0].Status != frontend.ParentPaymentNotInDatabase {
		t.Fatalf("status %q", rows[0].Status)
	}
}

func TestBuildParentPaymentVerificationRowsScanFailed(t *testing.T) {
	entries := []parentpaymentocr.PaymentOCREntry{{ScanFailed: true}}
	rows := buildParentPaymentVerificationRows(entries, nil, nil)
	if len(rows) != 1 || rows[0].Status != frontend.ParentPaymentOCRScanFailed {
		t.Fatalf("rows: %+v", rows)
	}
}
