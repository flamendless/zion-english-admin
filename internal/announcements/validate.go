package announcements

import (
	"errors"
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
)

type Request struct {
	Title         string
	Description   string
	Level         string
	StartDate     string
	EndDate       string
	VisibleToAll  bool
	TeacherIDs    []int64
	OriginalStart string
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

	if isUpdate && req.OriginalStart != "" {
		if req.OriginalStart < today {
			// Already started — keep original start even if in the past.
		} else if startDay < today {
			return ErrStartDatePast
		}
	} else if startDay < today {
		return ErrStartDatePast
	}

	if !req.VisibleToAll && len(req.TeacherIDs) == 0 {
		return ErrTeachersRequired
	}

	return nil
}
