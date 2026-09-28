package constants

type AdRandomizeKind string

const (
	AdRandomizePerPage    AdRandomizeKind = "per_page"
	AdRandomizePerSession AdRandomizeKind = "per_session"
	AdRandomizeTimer      AdRandomizeKind = "timer"
)

func ValidAdRandomizeKind(kind string) bool {
	switch AdRandomizeKind(kind) {
	case AdRandomizePerPage, AdRandomizePerSession, AdRandomizeTimer:
		return true
	default:
		return false
	}
}

type AdTimerInterval string

const (
	AdTimerIntervalFiveSeconds AdTimerInterval = "5_seconds"
	AdTimerIntervalHourly      AdTimerInterval = "hourly"
	AdTimerIntervalDaily       AdTimerInterval = "daily"
)

func ValidAdTimerInterval(interval string) bool {
	if interval == "" {
		return true
	}
	switch AdTimerInterval(interval) {
	case AdTimerIntervalFiveSeconds, AdTimerIntervalHourly, AdTimerIntervalDaily:
		return true
	default:
		return false
	}
}
