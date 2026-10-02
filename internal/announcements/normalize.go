package announcements

import "database/sql"

// NormalizeRequest clears modal-only fields for banners and applies defaults after validation.
func NormalizeRequest(req Request) Request {
	if DisplayType(req.DisplayType) != DisplayTypeModal {
		req.DisplayType = string(DisplayTypeBanner)
		req.ModalFrequency = ""
		req.RepeatEnabled = false
		req.RepeatSchedule = ""
		req.CutoffRepeatDays = sql.NullInt64{}
		return req
	}
	if !req.RepeatEnabled {
		req.RepeatSchedule = ""
		req.CutoffRepeatDays = sql.NullInt64{}
	} else if RepeatSchedule(req.RepeatSchedule) != RepeatScheduleCutoffBefore {
		req.CutoffRepeatDays = sql.NullInt64{}
	}
	return req
}
