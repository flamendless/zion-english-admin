package calendar

import "errors"

var (
	ErrGoogleCalendarNotConfigured = errors.New("[CALENDAR] google calendar integration is not configured")
	ErrGoogleCalendarNotConnected  = errors.New("[CALENDAR] connect google calendar on your profile first")
	ErrProviderNotFound            = errors.New("[CALENDAR] calendar provider not found")
	ErrStartTimeRequired           = errors.New("[CALENDAR] start time is required for calendar sync")
	ErrAuthorizationCodeRequired   = errors.New("[CALENDAR] authorization code is required")
)
