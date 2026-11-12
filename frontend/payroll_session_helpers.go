package frontend

import (
	"time"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

func (d PayrollSessionData) SessionMonth() string {
	t, err := time.ParseInLocation(constants.DateLayout, d.StartDate, constants.LocationPHT)
	if err != nil {
		return utils.CurrentMonthPHT()
	}
	return t.Format(constants.MonthLayout)
}

func (d PayrollSessionData) SessionPreset() string {
	first, second := utils.CutoffRangeForMonth(
		mustParseMonthYear(d.SessionMonth()),
		mustParseMonth(d.SessionMonth()),
	)
	if d.StartDate+"|"+d.EndDate == first {
		return first
	}
	if d.StartDate+"|"+d.EndDate == second {
		return second
	}
	return d.StartDate + "|" + d.EndDate
}

func (d PayrollSessionData) RoleTeacherWire() string {
	if d.RoleTeacher {
		return "1"
	}
	return "0"
}

func (d PayrollSessionData) RoleAdminWire() string {
	if d.RoleAdmin {
		return "1"
	}
	return "0"
}

func mustParseMonth(monthValue string) time.Month {
	_, m, ok := utils.ParseMonthPHT(monthValue)
	if !ok {
		return time.Now().In(constants.LocationPHT).Month()
	}
	return m
}

func mustParseMonthYear(monthValue string) int {
	y, _, ok := utils.ParseMonthPHT(monthValue)
	if !ok {
		return time.Now().In(constants.LocationPHT).Year()
	}
	return y
}

func (row PayrollSessionListItem) OpenLabel() string {
	if row.IsActive {
		return "Continue"
	}
	return "Open"
}
