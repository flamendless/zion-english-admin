package constants

type TeacherStatus string

const (
	TeacherStatusPending  TeacherStatus = "pending"
	TeacherStatusApproved TeacherStatus = "approved"
)

const TeacherDisplayStatusDeleted = "deleted"

type TeacherFilterStatus string

const (
	TeacherFilterStatusPending  TeacherFilterStatus = "pending"
	TeacherFilterStatusApproved TeacherFilterStatus = "approved"
	TeacherFilterStatusDeleted  TeacherFilterStatus = "deleted"
)

var TeacherFilterStatuses = []TeacherFilterStatus{
	TeacherFilterStatusPending,
	TeacherFilterStatusApproved,
	TeacherFilterStatusDeleted,
}

var TeacherStatuses = []TeacherStatus{
	TeacherStatusPending,
	TeacherStatusApproved,
}

func (s TeacherStatus) String() string {
	return string(s)
}

func (s TeacherFilterStatus) Label() string {
	switch s {
	case TeacherFilterStatusPending:
		return "Pending"
	case TeacherFilterStatusApproved:
		return "Approved"
	case TeacherFilterStatusDeleted:
		return "Deleted"
	default:
		return string(s)
	}
}

func ValidTeacherStatus(status string) bool {
	return status == string(TeacherStatusPending) || status == string(TeacherStatusApproved)
}
