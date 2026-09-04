package cmd

import (
	"testing"

	"zion-english/internal/constants"
	"zion-english/internal/models"
)

func TestValidateClassRecordRequestZeroesRateForCancelledOrRescheduled(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		rate       float64
		wantRate   float64
		trial      bool
		wantErr    bool
		errContain string
	}{
		{
			name:     "conducted requires positive rate",
			status:   string(constants.ClassStatusConducted),
			rate:     100,
			wantRate: 100,
		},
		{
			name:     "trial conducted applies 50 PHP",
			status:   string(constants.ClassStatusConducted),
			rate:     100,
			wantRate: constants.TrialClassRate,
			trial:    true,
		},
		{
			name:       "conducted rejects zero rate",
			status:     string(constants.ClassStatusConducted),
			rate:       0,
			wantRate:   0,
			wantErr:    true,
			errContain: "rate must be greater than zero",
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
			req := &models.ClassRecordRequest{
				StudentID:       1,
				TeacherID:       2,
				Date:            "2026-08-01",
				DurationMinutes: 60,
				Rate:            tt.rate,
				Currency:        "KRW",
				IsTrialClass:    tt.trial,
				Status:          tt.status,
				Reason:          "student unavailable",
			}
			if tt.status == string(constants.ClassStatusConducted) {
				req.Reason = ""
			}

			err := validateClassRecordRequest(req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("validateClassRecordRequest() error = nil, want error")
				}
				if tt.errContain != "" && err.Error() != tt.errContain {
					t.Fatalf("validateClassRecordRequest() error = %q, want %q", err.Error(), tt.errContain)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateClassRecordRequest() unexpected error: %v", err)
			}
			if req.Rate != tt.wantRate {
				t.Fatalf("req.Rate = %v, want %v", req.Rate, tt.wantRate)
			}
			if tt.trial && req.Currency != constants.TrialClassCurrency {
				t.Fatalf("req.Currency = %q, want %q", req.Currency, constants.TrialClassCurrency)
			}
		})
	}
}
