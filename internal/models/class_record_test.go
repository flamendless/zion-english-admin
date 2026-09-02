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
