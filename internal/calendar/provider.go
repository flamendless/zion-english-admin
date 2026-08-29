package calendar

import "context"

type AccountRef struct {
	TeacherID      int64
	Service        string
	ExternalUserID string
	ResourceID     string
	AccessToken    string
	RefreshToken   string
}

type CreateEventRequest struct {
	Summary         string
	Description     string
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
}

type UpdateEventRequest struct {
	Summary         string
	Description     string
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
}

type Event struct {
	Service  string
	EventID  string
	EventURL string
}

type Provider interface {
	Service() string
	EnsureDedicatedCalendar(ctx context.Context, account AccountRef) (calendarID string, err error)
	CreateEvent(ctx context.Context, account AccountRef, req CreateEventRequest) (Event, error)
	UpdateEvent(ctx context.Context, account AccountRef, eventID string, req UpdateEventRequest) (Event, error)
	DeleteEvent(ctx context.Context, account AccountRef, eventID string) error
}
