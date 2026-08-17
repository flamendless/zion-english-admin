package announcements_test

import (
	"testing"
	"time"

	"zion-english/internal/announcements"
	"zion-english/internal/constants"
)

func todayString() string {
	now := time.Now().In(constants.LocationPHT)
	return now.Format(constants.DateLayout)
}

func tomorrowString() string {
	now := time.Now().In(constants.LocationPHT)
	return now.AddDate(0, 0, 1).Format(constants.DateLayout)
}

func dayAfterTomorrowString() string {
	now := time.Now().In(constants.LocationPHT)
	return now.AddDate(0, 0, 2).Format(constants.DateLayout)
}

func TestValidateRequestCreate(t *testing.T) {
	today := todayString()
	tomorrow := tomorrowString()
	dayAfter := dayAfterTomorrowString()

	tests := []struct {
		name    string
		req     announcements.Request
		wantErr error
	}{
		{
			name: "valid all teachers",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
			},
		},
		{
			name: "missing title",
			req: announcements.Request{
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
			},
			wantErr: announcements.ErrTitleRequired,
		},
		{
			name: "past start date",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    "2020-01-01",
				EndDate:      tomorrow,
				VisibleToAll: true,
			},
			wantErr: announcements.ErrStartDatePast,
		},
		{
			name: "end before start",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    dayAfter,
				EndDate:      tomorrow,
				VisibleToAll: true,
			},
			wantErr: announcements.ErrEndBeforeStart,
		},
		{
			name: "selected teachers required",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelWarning,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: false,
			},
			wantErr: announcements.ErrTeachersRequired,
		},
		{
			name: "cta label without url",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
				CTALabel:     "Learn more",
			},
			wantErr: announcements.ErrCTAURLRequired,
		},
		{
			name: "cta url without label",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
				CTAURL:       "https://example.com",
			},
			wantErr: announcements.ErrCTALabelRequired,
		},
		{
			name: "invalid cta url",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
				CTALabel:     "Go",
				CTAURL:       "javascript:alert(1)",
			},
			wantErr: announcements.ErrInvalidCTAURL,
		},
		{
			name: "valid cta external url",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
				CTALabel:     "Learn more",
				CTAURL:       "https://example.com",
			},
		},
		{
			name: "valid cta internal path",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
				CTALabel:     "View students",
				CTAURL:       "/students",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := announcements.ValidateRequest(tt.req, false)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
