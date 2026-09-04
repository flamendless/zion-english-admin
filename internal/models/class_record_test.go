package models

import (
	"testing"

	"zion-english/internal/constants"
)

func TestNormalizeClassRecordRate(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		rate     float64
		wantRate float64
	}{
		{
			name:     "conducted keeps rate",
			status:   string(constants.ClassStatusConducted),
			rate:     250,
			wantRate: 250,
		},
		{
			name:     "cancelled zeroes rate",
			status:   string(constants.ClassStatusCancelled),
			rate:     250,
			wantRate: 0,
		},
		{
			name:     "rescheduled zeroes rate",
			status:   string(constants.ClassStatusRescheduled),
			rate:     180,
			wantRate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ClassRecordRequest{Status: tt.status, Rate: tt.rate}
			NormalizeClassRecordRate(req)
			if req.Rate != tt.wantRate {
				t.Fatalf("Rate = %v, want %v", req.Rate, tt.wantRate)
			}
		})
	}
}

func TestApplyTrialClassRate(t *testing.T) {
	req := &ClassRecordRequest{
		Rate:     300,
		Currency: "KRW",
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

func TestApplyTrialClassRateThenNormalizeCancelled(t *testing.T) {
	req := &ClassRecordRequest{
		Status:       string(constants.ClassStatusCancelled),
		Rate:         300,
		Currency:     "KRW",
		IsTrialClass: true,
	}
	ApplyTrialClassRate(req)
	NormalizeClassRecordRate(req)
	if req.Rate != 0 {
		t.Fatalf("Rate = %v, want 0", req.Rate)
	}
}
