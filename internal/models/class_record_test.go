package models

import (
	"testing"

	"zion-english/internal/constants"
)

func TestApplyTrialClassRate(t *testing.T) {
	req := &ClassRecordRequest{
		Rate:         300,
		Currency:     "KRW",
		IsTrialClass: true,
	}
	ApplyTrialClassRate(req)
	if req.Rate != constants.TrialClassRate {
		t.Fatalf("Rate = %v, want %v", req.Rate, constants.TrialClassRate)
	}
	if req.Currency != constants.TrialClassCurrency {
		t.Fatalf("Currency = %q, want %q", req.Currency, constants.TrialClassCurrency)
	}
}

func TestApplyTrialClassRatePreservesRateForCancelled(t *testing.T) {
	req := &ClassRecordRequest{
		Status:       string(constants.ClassStatusCancelled),
		Rate:         300,
		Currency:     "KRW",
		IsTrialClass: true,
	}
	ApplyTrialClassRate(req)
	if req.Rate != constants.TrialClassRate {
		t.Fatalf("Rate = %v, want %v", req.Rate, constants.TrialClassRate)
	}
}
