package utils

import "errors"

var (
	ErrDriveURLRequired            = errors.New("[DRIVE] spreadsheet URL is required")
	ErrInvalidDriveSpreadsheetURL  = errors.New("[DRIVE] invalid spreadsheet URL")
	ErrInvalidDriveSpreadsheetPath = errors.New("[DRIVE] invalid Google Sheets URL path: expected /spreadsheets/d/{DOCUMENT_ID}/.../")

	ErrTimeRequired      = errors.New("time is required")
	ErrInvalidTimeFormat = errors.New("invalid time format")
	ErrInvalidStartTime  = errors.New("invalid start time")
	ErrInvalidEndTime    = errors.New("invalid end time")
	ErrEndBeforeStart    = errors.New("end time must be after start time")
)
