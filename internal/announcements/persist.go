package announcements

import (
	"database/sql"
	"strconv"
	"strings"
)

// DBFields holds announcement columns stored in tbl_announcements.
type DBFields struct {
	DisplayType      string
	ModalFrequency   sql.NullString
	RepeatEnabled    int64
	RepeatSchedule   sql.NullString
	CutoffRepeatDays sql.NullInt64
}

func DBFieldsFromRequest(req Request) DBFields {
	req = NormalizeRequest(req)
	out := DBFields{
		DisplayType:   string(DisplayTypeBanner),
		RepeatEnabled: 0,
	}
	if DisplayType(req.DisplayType) == DisplayTypeModal {
		out.DisplayType = string(DisplayTypeModal)
		if req.ModalFrequency != "" {
			out.ModalFrequency = sql.NullString{String: req.ModalFrequency, Valid: true}
		}
		if req.RepeatEnabled {
			out.RepeatEnabled = 1
			if req.RepeatSchedule != "" {
				out.RepeatSchedule = sql.NullString{String: req.RepeatSchedule, Valid: true}
			}
			out.CutoffRepeatDays = req.CutoffRepeatDays
		}
	}
	return out
}

// ParseCutoffRepeatDaysFromForm parses the optional cutoff repeat days field.
func ParseCutoffRepeatDaysFromForm(raw string) (sql.NullInt64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sql.NullInt64{}, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return sql.NullInt64{}, ErrInvalidCutoffRepeatDays
	}
	if !ValidCutoffRepeatDays(n) {
		return sql.NullInt64{}, ErrInvalidCutoffRepeatDays
	}
	return sql.NullInt64{Int64: int64(n), Valid: true}, nil
}
