package announcements

import (
	"database/sql"

	"zion-english/internal/utils"
)

// VisibilityInput carries schedule fields used to decide if an announcement shows on a day.
type VisibilityInput struct {
	StartDate        string
	EndDate          string
	RepeatEnabled    bool
	RepeatSchedule   RepeatSchedule
	CutoffRepeatDays sql.NullInt64
}

func IsVisibleOnDate(in VisibilityInput, today string) bool {
	if today < in.StartDate || today > in.EndDate {
		return false
	}
	if !in.RepeatEnabled {
		return true
	}
	switch in.RepeatSchedule {
	case RepeatScheduleStartEnd:
		return true
	case RepeatScheduleCutoffBefore:
		if !in.CutoffRepeatDays.Valid {
			return false
		}
		n := int(in.CutoffRepeatDays.Int64)
		return utils.InCutoffRepeatWindow(today, n)
	default:
		return false
	}
}

func VisibilityInputFromStorage(
	startDate, endDate string,
	repeatEnabled int64,
	repeatSchedule sql.NullString,
	cutoffRepeatDays sql.NullInt64,
) VisibilityInput {
	in := VisibilityInput{
		StartDate:        startDate,
		EndDate:          endDate,
		RepeatEnabled:    repeatEnabled != 0,
		CutoffRepeatDays: cutoffRepeatDays,
	}
	if repeatSchedule.Valid {
		in.RepeatSchedule = RepeatSchedule(repeatSchedule.String)
	}
	return in
}
