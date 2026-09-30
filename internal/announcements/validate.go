package announcements

import (
	"database/sql"
	"net/url"
	"strings"

	"zion-english/internal/utils"
)

const maxCTALabelLen = 60

type Request struct {
	Title            string
	Description      string
	Level            string
	StartDate        string
	EndDate          string
	VisibleToAll     bool
	TeacherIDs       []int64
	OriginalStart    string
	CTALabel         string
	CTAURL           string
	Status           string
	DisplayType      string
	ModalFrequency   string
	RepeatEnabled    bool
	RepeatSchedule   string
	CutoffRepeatDays sql.NullInt64
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
	if !ValidFormStatus(req.Status) {
		return ErrInvalidStatus
	}
	displayType := req.DisplayType
	if displayType == "" {
		displayType = string(DisplayTypeBanner)
	}
	if !ValidDisplayType(displayType) {
		return ErrInvalidDisplayType
	}
	if DisplayType(displayType) == DisplayTypeModal {
		if !ValidModalFrequency(req.ModalFrequency) {
			if strings.TrimSpace(req.ModalFrequency) == "" {
				return ErrModalFrequencyRequired
			}
			return ErrInvalidModalFrequency
		}
		if req.RepeatEnabled {
			if !ValidRepeatSchedule(req.RepeatSchedule) {
				if strings.TrimSpace(req.RepeatSchedule) == "" {
					return ErrRepeatScheduleRequired
				}
				return ErrInvalidRepeatSchedule
			}
			if RepeatSchedule(req.RepeatSchedule) == RepeatScheduleCutoffBefore {
				if !req.CutoffRepeatDays.Valid {
					return ErrCutoffRepeatDaysRequired
				}
				if !ValidCutoffRepeatDays(int(req.CutoffRepeatDays.Int64)) {
					return ErrInvalidCutoffRepeatDays
				}
			}
		}
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
