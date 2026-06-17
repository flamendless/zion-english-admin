package constants

type TeacherStatus string

const (
	TeacherStatusPending  TeacherStatus = "pending"
	TeacherStatusApproved TeacherStatus = "approved"
)

func (s TeacherStatus) String() string {
	return string(s)
}

func ValidTeacherStatus(status string) bool {
	return status == string(TeacherStatusPending) || status == string(TeacherStatusApproved)
}
