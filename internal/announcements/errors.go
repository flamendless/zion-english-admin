package announcements

import "errors"

var (
	ErrTitleRequired       = errors.New("[ANNOUNCEMENTS] title is required")
	ErrDescriptionRequired = errors.New("[ANNOUNCEMENTS] description is required")
	ErrInvalidLevel        = errors.New("[ANNOUNCEMENTS] invalid announcement level")
	ErrStartDateRequired   = errors.New("[ANNOUNCEMENTS] start date is required")
	ErrEndDateRequired     = errors.New("[ANNOUNCEMENTS] end date is required")
	ErrInvalidStartDate    = errors.New("[ANNOUNCEMENTS] invalid start date format")
	ErrInvalidEndDate      = errors.New("[ANNOUNCEMENTS] invalid end date format")
	ErrStartDatePast       = errors.New("[ANNOUNCEMENTS] start date cannot be in the past")
	ErrEndDatePast         = errors.New("[ANNOUNCEMENTS] end date cannot be in the past")
	ErrEndBeforeStart      = errors.New("[ANNOUNCEMENTS] end date must be on or after start date")
	ErrTeachersRequired    = errors.New("[ANNOUNCEMENTS] select at least one teacher when not visible to all")
	ErrCTALabelRequired    = errors.New("[ANNOUNCEMENTS] CTA label is required when CTA URL is set")
	ErrCTAURLRequired      = errors.New("[ANNOUNCEMENTS] CTA URL is required when CTA label is set")
	ErrCTALabelTooLong     = errors.New("[ANNOUNCEMENTS] CTA label must be 60 characters or fewer")
	ErrInvalidCTAURL       = errors.New("[ANNOUNCEMENTS] CTA URL must be http(s):// or an internal path starting with /")
	ErrInvalidStatus       = errors.New("[ANNOUNCEMENTS] invalid announcement status")
)
