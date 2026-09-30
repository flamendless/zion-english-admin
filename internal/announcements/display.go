package announcements

type DisplayType string

const (
	DisplayTypeBanner DisplayType = "banner"
	DisplayTypeModal  DisplayType = "modal"
)

func ValidDisplayType(value string) bool {
	switch DisplayType(value) {
	case DisplayTypeBanner, DisplayTypeModal:
		return true
	default:
		return false
	}
}

type ModalFrequency string

const (
	ModalFrequencyPerSession ModalFrequency = "per_session"
	ModalFrequencyPerPage    ModalFrequency = "per_page"
)

func ValidModalFrequency(value string) bool {
	switch ModalFrequency(value) {
	case ModalFrequencyPerSession, ModalFrequencyPerPage:
		return true
	default:
		return false
	}
}

type RepeatSchedule string

const (
	RepeatScheduleStartEnd     RepeatSchedule = "start_end"
	RepeatScheduleCutoffBefore RepeatSchedule = "cutoff_before"
)

func ValidRepeatSchedule(value string) bool {
	switch RepeatSchedule(value) {
	case RepeatScheduleStartEnd, RepeatScheduleCutoffBefore:
		return true
	default:
		return false
	}
}

const (
	minCutoffRepeatDays = 0
	maxCutoffRepeatDays = 10
)

func ValidCutoffRepeatDays(n int) bool {
	return n >= minCutoffRepeatDays && n <= maxCutoffRepeatDays
}
