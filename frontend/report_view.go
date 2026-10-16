package frontend

import (
	"fmt"

	"zion-english/internal/constants"
)

func formatReportClassDuration(durationMinutes int64) string {
	if durationMinutes <= 0 {
		return constants.DisplayNotSet
	}
	return fmt.Sprintf("%d", durationMinutes)
}
