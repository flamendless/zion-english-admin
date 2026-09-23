package teacherintrovideo

import "errors"

var (
	ErrURLRequired           = errors.New("[INTRO VIDEO] url is required")
	ErrInvalidURL            = errors.New("[INTRO VIDEO] url must be a valid link for the selected source")
	ErrUnsupportedSourceType = errors.New("[INTRO VIDEO] source type must be upload, google drive, or youtube")
	ErrFileEmpty             = errors.New("[INTRO VIDEO] Uploaded file is empty")
	ErrFileTooLarge          = errors.New("[INTRO VIDEO] File is too large. Maximum size is 20 MB.")
	ErrUnsupportedFormat     = errors.New("[INTRO VIDEO] Unsupported file format. Please upload MP4, WebM, MOV, AVI, MKV, OGV, M4V, or 3GP.")
	ErrUploadPrepareFailed   = errors.New("[INTRO VIDEO] Failed to prepare upload")
	ErrReadFailed            = errors.New("[INTRO VIDEO] Failed to read uploaded file")
	ErrInvalidContent        = errors.New("[INTRO VIDEO] Invalid video file. Please upload a valid video.")
	ErrEmptyDuration         = errors.New("[INTRO VIDEO] empty duration")
	ErrTooLong               = errors.New("[INTRO VIDEO] Video is too long. Maximum duration is 60 seconds.")
	ErrFfprobeUnavailable    = errors.New("[INTRO VIDEO] ffprobe unavailable")
)
