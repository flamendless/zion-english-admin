package frontend

import (
	"strconv"

	"zion-english/internal/constants"
)

func PayrollTeacherStatusPillTone(status constants.PayrollTeacherStatus) PillTone {
	switch status {
	case constants.PayrollTeacherStatusAccomplishDailyLogs:
		return PillToneWarning
	case constants.PayrollTeacherStatusDoneDailyLogs:
		return PillToneInfo
	case constants.PayrollTeacherStatusAudited:
		return PillTonePrimary
	case constants.PayrollTeacherStatusSentSalary:
		return PillToneWarning
	case constants.PayrollTeacherStatusSalaryReceived:
		return PillToneSuccess
	default:
		return PillToneNeutral
	}
}

func PayrollTeacherStatusLabel(status constants.PayrollTeacherStatus) string {
	if status.Step() == 0 {
		return status.Label()
	}
	return "Step " + strconv.Itoa(status.Step()) + " · " + status.Label()
}

func PayrollSessionStatusPillTone(status constants.PayrollSessionStatus) PillTone {
	switch status {
	case constants.PayrollSessionStatusActive:
		return PillToneWarning
	case constants.PayrollSessionStatusFinished:
		return PillToneSuccess
	default:
		return PillToneNeutral
	}
}

var PayrollTeacherStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.PayrollTeacherStatuses))
	for _, status := range constants.PayrollTeacherStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()
