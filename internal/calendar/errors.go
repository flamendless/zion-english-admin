package calendar

import "errors"

var (
	ErrGoogleCalendarNotConfigured = errors.New("google calendar integration is not configured")
	ErrGoogleCalendarNotConnected  = errors.New("connect google calendar on your profile first")
	ErrProviderNotFound            = errors.New("calendar provider not found")
	ErrStartTimeRequired           = errors.New("start time is required for calendar sync")
)
