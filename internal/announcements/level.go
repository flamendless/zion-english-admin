package announcements

const (
	LevelInfo     = "info"
	LevelWarning  = "warning"
	LevelCritical = "critical"
)

func ValidLevel(level string) bool {
	switch level {
	case LevelInfo, LevelWarning, LevelCritical:
		return true
	default:
		return false
	}
}

func AllLevels() []string {
	return []string{LevelInfo, LevelWarning, LevelCritical}
}
