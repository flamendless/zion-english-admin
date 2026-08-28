package constants

type StudentStatus string

const (
	StudentStatusActive   StudentStatus = "active"
	StudentStatusInactive StudentStatus = "inactive"
)

var StudentStatuses = []StudentStatus{
	StudentStatusActive,
	StudentStatusInactive,
}

func (s StudentStatus) String() string {
	return string(s)
}

func (s StudentStatus) Label() string {
	switch s {
	case StudentStatusActive:
		return "Active"
	case StudentStatusInactive:
		return "Inactive"
	default:
		return string(s)
	}
}

func ValidStudentStatus(status string) bool {
	switch StudentStatus(status) {
	case StudentStatusActive, StudentStatusInactive:
		return true
	default:
		return false
	}
}
