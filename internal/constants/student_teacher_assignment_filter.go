package constants

type StudentTeacherAssignmentFilter string

const (
	StudentTeacherAssignmentFilterNoApproved StudentTeacherAssignmentFilter = "no_approved_teacher"
)

func ValidStudentTeacherAssignmentFilter(value string) bool {
	switch StudentTeacherAssignmentFilter(value) {
	case StudentTeacherAssignmentFilterNoApproved:
		return true
	default:
		return false
	}
}

func (f StudentTeacherAssignmentFilter) Label() string {
	switch f {
	case StudentTeacherAssignmentFilterNoApproved:
		return "Without valid teacher"
	default:
		return ""
	}
}
