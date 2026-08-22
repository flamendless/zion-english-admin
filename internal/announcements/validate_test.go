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

func yesterdayString() string {
	now := time.Now().In(constants.LocationPHT)
	return now.AddDate(0, 0, -1).Format(constants.DateLayout)
}

func validAnnouncementReq() announcements.Request {
	return announcements.Request{
		Title:        "Test",
		Description:  "Details",
		Level:        announcements.LevelInfo,
		StartDate:    todayString(),
		EndDate:      tomorrowString(),
		VisibleToAll: true,
		Status:       announcements.StatusDraft,
	}
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
				Status:       announcements.StatusDraft,
			},
		},
		{
			name: "missing title",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.Title = ""
				return req
			}(),
			wantErr: announcements.ErrTitleRequired,
		},
		{
			name: "past start date",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.StartDate = "2020-01-01"
				return req
			}(),
			wantErr: announcements.ErrStartDatePast,
		},
		{
			name: "end before start",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.StartDate = dayAfter
				req.EndDate = tomorrow
				return req
			}(),
			wantErr: announcements.ErrEndBeforeStart,
		},
		{
			name: "selected teachers required",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.Level = announcements.LevelWarning
				req.VisibleToAll = false
				return req
			}(),
			wantErr: announcements.ErrTeachersRequired,
		},
		{
			name: "cta label without url",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.CTALabel = "Learn more"
				return req
			}(),
			wantErr: announcements.ErrCTAURLRequired,
		},
		{
			name: "cta url without label",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.CTAURL = "https://example.com"
				return req
			}(),
			wantErr: announcements.ErrCTALabelRequired,
		},
		{
			name: "invalid cta url",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.CTALabel = "Go"
				req.CTAURL = "javascript:alert(1)"
				return req
			}(),
			wantErr: announcements.ErrInvalidCTAURL,
		},
		{
			name: "valid cta external url",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.CTALabel = "Learn more"
				req.CTAURL = "https://example.com"
				return req
			}(),
		},
		{
			name: "invalid status",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
				Status:       announcements.StatusDeleted,
			},
			wantErr: announcements.ErrInvalidStatus,
		},
		{
			name: "valid published",
			req: announcements.Request{
				Title:        "Test",
				Description:  "Details",
				Level:        announcements.LevelInfo,
				StartDate:    today,
				EndDate:      tomorrow,
				VisibleToAll: true,
				Status:       announcements.StatusPublished,
			},
		},
		{
			name: "valid cta internal path",
			req: func() announcements.Request {
				req := validAnnouncementReq()
				req.CTALabel = "View students"
				req.CTAURL = "/students"
				return req
			}(),
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

func TestValidateRequestUpdate(t *testing.T) {
	today := todayString()
	tomorrow := tomorrowString()
	yesterday := yesterdayString()

	tests := []struct {
		name    string
		req     announcements.Request
		wantErr error
	}{
		{
			name: "unchanged past start date",
			req: announcements.Request{
				Title:         "Test",
				Description:   "Details",
				Level:         announcements.LevelInfo,
				StartDate:     yesterday,
				EndDate:       tomorrow,
				OriginalStart: yesterday,
				VisibleToAll:  true,
				Status:        announcements.StatusDraft,
			},
		},
		{
			name: "changed start to past",
			req: announcements.Request{
				Title:         "Test",
				Description:   "Details",
				Level:         announcements.LevelInfo,
				StartDate:     yesterday,
				EndDate:       tomorrow,
				OriginalStart: today,
				VisibleToAll:  true,
				Status:        announcements.StatusDraft,
			},
			wantErr: announcements.ErrStartDatePast,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := announcements.ValidateRequest(tt.req, true)
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
