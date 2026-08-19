package announcements

import (
	"errors"
	"net/url"
	"strings"

	"zion-english/internal/utils"
)

var (
	ErrTitleRequired       = errors.New("title is required")
	ErrDescriptionRequired = errors.New("description is required")
	ErrInvalidLevel        = errors.New("invalid announcement level")
	ErrStartDateRequired   = errors.New("start date is required")
	ErrEndDateRequired     = errors.New("end date is required")
	ErrInvalidStartDate    = errors.New("invalid start date format")
	ErrInvalidEndDate      = errors.New("invalid end date format")
	ErrStartDatePast       = errors.New("start date cannot be in the past")
	ErrEndDatePast         = errors.New("end date cannot be in the past")
	ErrEndBeforeStart      = errors.New("end date must be on or after start date")
	ErrTeachersRequired    = errors.New("select at least one teacher when not visible to all")
	ErrCTALabelRequired    = errors.New("CTA label is required when CTA URL is set")
	ErrCTAURLRequired      = errors.New("CTA URL is required when CTA label is set")
	ErrCTALabelTooLong     = errors.New("CTA label must be 60 characters or fewer")
	ErrInvalidCTAURL       = errors.New("CTA URL must be http(s):// or an internal path starting with /")
)

const maxCTALabelLen = 60

type Request struct {
	Title         string
	Description   string
	Level         string
	StartDate     string
	EndDate       string
	VisibleToAll  bool
	TeacherIDs    []int64
	OriginalStart string
	CTALabel      string
	CTAURL        string
}

func ValidateRequest(req Request, isUpdate bool) error {
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if utils.IsBlank(req.Title) {
		return ErrTitleRequired
	}
	if utils.IsBlank(req.Description) {
		return ErrDescriptionRequired
	}
	if !ValidLevel(req.Level) {
		return ErrInvalidLevel
	}
	if req.StartDate == "" {
		return ErrStartDateRequired
	}
	if req.EndDate == "" {
		return ErrEndDateRequired
	}

	start, err := utils.ParseDatePHT(req.StartDate)
	if err != nil || start == nil {
		return ErrInvalidStartDate
	}
	end, err := utils.ParseDatePHT(req.EndDate)
	if err != nil || end == nil {
		return ErrInvalidEndDate
	}

	today := utils.TodayPHT()
	startDay := utils.DatePHT(*start)
	endDay := utils.DatePHT(*end)

	if endDay < startDay {
		return ErrEndBeforeStart
	}
	if endDay < today {
		return ErrEndDatePast
	}

	startUnchanged := isUpdate && req.OriginalStart != "" && req.StartDate == req.OriginalStart
	if !startUnchanged && startDay < today {
		return ErrStartDatePast
	}

	if !req.VisibleToAll && len(req.TeacherIDs) == 0 {
		return ErrTeachersRequired
	}

	req.CTALabel = strings.TrimSpace(req.CTALabel)
	req.CTAURL = strings.TrimSpace(req.CTAURL)
	if req.CTALabel == "" && req.CTAURL == "" {
		return nil
	}
	if req.CTALabel == "" {
		return ErrCTALabelRequired
	}
	if req.CTAURL == "" {
		return ErrCTAURLRequired
	}
	if len(req.CTALabel) > maxCTALabelLen {
		return ErrCTALabelTooLong
	}
	if !validCTAURL(req.CTAURL) {
		return ErrInvalidCTAURL
	}

	return nil
}

func validCTAURL(raw string) bool {
	if strings.HasPrefix(raw, "/") {
		return true
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
