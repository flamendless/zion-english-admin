package constants

type TeacherStatus string

const (
	TeacherStatusPending  TeacherStatus = "pending"
	TeacherStatusApproved TeacherStatus = "approved"
	TeacherStatusResigned TeacherStatus = "resigned"
)

const TeacherDisplayStatusDeleted = "deleted"

type TeacherFilterStatus string

const (
	TeacherFilterStatusPending  TeacherFilterStatus = "pending"
	TeacherFilterStatusApproved TeacherFilterStatus = "approved"
	TeacherFilterStatusResigned TeacherFilterStatus = "resigned"
	TeacherFilterStatusDeleted  TeacherFilterStatus = "deleted"
)

var TeacherFilterStatuses = []TeacherFilterStatus{
	TeacherFilterStatusPending,
	TeacherFilterStatusApproved,
	TeacherFilterStatusResigned,
	TeacherFilterStatusDeleted,
}

var TeacherStatuses = []TeacherStatus{
	TeacherStatusPending,
	TeacherStatusApproved,
	TeacherStatusResigned,
}

func (s TeacherStatus) String() string {
	return string(s)
}

func (s TeacherStatus) Label() string {
	switch s {
	case TeacherStatusPending:
		return "Pending"
	case TeacherStatusApproved:
		return "Approved"
	case TeacherStatusResigned:
		return "Resigned"
	default:
		return string(s)
	}
}

func (s TeacherFilterStatus) Label() string {
	switch s {
	case TeacherFilterStatusPending:
		return "Pending"
	case TeacherFilterStatusApproved:
		return "Approved"
	case TeacherFilterStatusResigned:
		return "Resigned"
	case TeacherFilterStatusDeleted:
		return "Deleted"
	default:
		return string(s)
	}
}

func ValidTeacherStatus(status string) bool {
	switch TeacherStatus(status) {
	case TeacherStatusPending, TeacherStatusApproved, TeacherStatusResigned:
		return true
	default:
		return false
	}
}

func ValidTeacherFilterStatus(status string) bool {
	if status == "" {
		return true
	}
	switch TeacherFilterStatus(status) {
	case TeacherFilterStatusPending, TeacherFilterStatusApproved, TeacherFilterStatusResigned, TeacherFilterStatusDeleted:
		return true
	default:
		return false
	}
}
