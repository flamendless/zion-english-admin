package constants

type PayrollTeacherStatus string

const (
	PayrollTeacherStatusAccomplishDailyLogs PayrollTeacherStatus = "accomplish_daily_logs"
	PayrollTeacherStatusDoneDailyLogs       PayrollTeacherStatus = "done_daily_logs"
	PayrollTeacherStatusAudited             PayrollTeacherStatus = "audited"
	PayrollTeacherStatusSentSalary          PayrollTeacherStatus = "sent_salary"
	PayrollTeacherStatusSalaryReceived      PayrollTeacherStatus = "salary_received"
)

func (s PayrollTeacherStatus) String() string {
	return string(s)
}

func (s PayrollTeacherStatus) Step() int {
	switch s {
	case PayrollTeacherStatusAccomplishDailyLogs:
		return 1
	case PayrollTeacherStatusDoneDailyLogs:
		return 2
	case PayrollTeacherStatusAudited:
		return 3
	case PayrollTeacherStatusSentSalary:
		return 4
	case PayrollTeacherStatusSalaryReceived:
		return 5
	default:
		return 0
	}
}

func (s PayrollTeacherStatus) Label() string {
	switch s {
	case PayrollTeacherStatusAccomplishDailyLogs:
		return "Accomplish daily logs"
	case PayrollTeacherStatusDoneDailyLogs:
		return "Done daily logs"
	case PayrollTeacherStatusAudited:
		return "Audited"
	case PayrollTeacherStatusSentSalary:
		return "Sent salary"
	case PayrollTeacherStatusSalaryReceived:
		return "Salary received"
	default:
		return string(s)
	}
}

func ValidPayrollTeacherStatus(status string) bool {
	switch PayrollTeacherStatus(status) {
	case PayrollTeacherStatusAccomplishDailyLogs,
		PayrollTeacherStatusDoneDailyLogs,
		PayrollTeacherStatusAudited,
		PayrollTeacherStatusSentSalary,
		PayrollTeacherStatusSalaryReceived:
		return true
	default:
		return false
	}
}

var PayrollTeacherStatuses = []PayrollTeacherStatus{
	PayrollTeacherStatusAccomplishDailyLogs,
	PayrollTeacherStatusDoneDailyLogs,
	PayrollTeacherStatusAudited,
	PayrollTeacherStatusSentSalary,
	PayrollTeacherStatusSalaryReceived,
}
