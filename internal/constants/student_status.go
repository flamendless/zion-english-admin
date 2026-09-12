package constants

type StudentStatus string

const (
	StudentStatusActive   StudentStatus = "active"
	StudentStatusInactive StudentStatus = "inactive"
	StudentStatusDeleted  StudentStatus = "deleted"
)

var StudentStatuses = []StudentStatus{
	StudentStatusActive,
	StudentStatusInactive,
}

var StudentFilterStatuses = []StudentStatus{
	StudentStatusActive,
	StudentStatusInactive,
	StudentStatusDeleted,
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
	case StudentStatusDeleted:
		return "Deleted"
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

func ValidStudentFilterStatus(status string) bool {
	if status == "" {
		return true
	}
	switch StudentStatus(status) {
	case StudentStatusActive, StudentStatusInactive, StudentStatusDeleted:
		return true
	default:
		return false
	}
}
