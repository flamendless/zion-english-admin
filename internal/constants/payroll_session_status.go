package constants

type PayrollSessionStatus string

const (
	PayrollSessionStatusActive   PayrollSessionStatus = "active"
	PayrollSessionStatusFinished PayrollSessionStatus = "finished"
)

func (s PayrollSessionStatus) String() string {
	return string(s)
}

func (s PayrollSessionStatus) Label() string {
	switch s {
	case PayrollSessionStatusActive:
		return "Active"
	case PayrollSessionStatusFinished:
		return "Finished"
	default:
		return string(s)
	}
}

func ValidPayrollSessionStatus(status string) bool {
	switch PayrollSessionStatus(status) {
	case PayrollSessionStatusActive, PayrollSessionStatusFinished:
		return true
	default:
		return false
	}
}
