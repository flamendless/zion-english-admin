package cmd

import (
	"database/sql"
	"testing"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
)

func TestEffectiveSummaryParentRate(t *testing.T) {
	parent := sql.NullFloat64{Float64: 100, Valid: true}
	currency := sql.NullString{String: "KRW", Valid: true}

	display, amount, ok := effectiveSummaryParentRate(parent, currency, constants.ClassStatusConducted, 100, 200)
	if !ok || amount != 100 || display != "100.00 KRW" {
		t.Fatalf("conducted: display=%q amount=%v ok=%v", display, amount, ok)
	}

	display, amount, ok = effectiveSummaryParentRate(parent, currency, constants.ClassStatusCancelled, 50, 100)
	if !ok || amount != 50 || display != "50.00 KRW" {
		t.Fatalf("cancelled half class rate: display=%q amount=%v ok=%v", display, amount, ok)
	}

	display, amount, ok = effectiveSummaryParentRate(parent, currency, constants.ClassStatusCancelled, 0, 100)
	if !ok || amount != 0 || display != "0.00 KRW" {
		t.Fatalf("cancelled zero class rate: display=%q amount=%v want 0", display, amount)
	}

	display, amount, ok = effectiveSummaryParentRate(parent, currency, constants.ClassStatusRescheduled, 40, 80)
	if !ok || amount != 50 {
		t.Fatalf("rescheduled half class rate: amount=%v", amount)
	}

	display, amount, ok = effectiveSummaryParentRate(parent, currency, constants.ClassStatusRescheduled, 0, 80)
	if !ok || amount != 0 {
		t.Fatalf("rescheduled zero class rate: amount=%v want 0", amount)
	}
}

func TestSummaryClassRowFromRecord(t *testing.T) {
	row := queries.GetReportSummaryRowsRow{
		Date:                "2026-08-15",
		Status:              string(constants.ClassStatusCancelled),
		Rate:                0,
		Currency:            "KRW",
		TeacherRatePerClass: 200,
		ParentRate:          sql.NullFloat64{Float64: 90, Valid: true},
		ParentCurrency:      sql.NullString{String: "KRW", Valid: true},
	}
	classRow := summaryClassRowFromRecord(row)
	if classRow.RateValue != 0 {
		t.Fatalf("cancelled with zero class rate: parent bill = %v, want 0", classRow.RateValue)
	}

	row.Rate = 100
	classRow = summaryClassRowFromRecord(row)
	if classRow.RateValue != 45 {
		t.Fatalf("cancelled with half class rate: parent bill = %v, want 45", classRow.RateValue)
	}

	row.Status = string(constants.ClassStatusConducted)
	classRow = summaryClassRowFromRecord(row)
	if classRow.RateValue != 90 {
		t.Fatalf("conducted: parent bill = %v, want 90", classRow.RateValue)
	}
}
