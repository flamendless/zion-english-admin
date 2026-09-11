package constants

type TeacherIntroVideoStatus string

const (
	TeacherIntroVideoStatusSubmitted TeacherIntroVideoStatus = "submitted"
	TeacherIntroVideoStatusApproved  TeacherIntroVideoStatus = "approved"
	TeacherIntroVideoStatusRejected  TeacherIntroVideoStatus = "rejected"
	TeacherIntroVideoStatusDeleted   TeacherIntroVideoStatus = "deleted"
)

var TeacherIntroVideoStatuses = []TeacherIntroVideoStatus{
	TeacherIntroVideoStatusSubmitted,
	TeacherIntroVideoStatusApproved,
	TeacherIntroVideoStatusRejected,
	TeacherIntroVideoStatusDeleted,
}

func (s TeacherIntroVideoStatus) Label() string {
	switch s {
	case TeacherIntroVideoStatusSubmitted:
		return "Submitted"
	case TeacherIntroVideoStatusApproved:
		return "Approved"
	case TeacherIntroVideoStatusRejected:
		return "Rejected"
	case TeacherIntroVideoStatusDeleted:
		return "Deleted"
	default:
		return string(s)
	}
}

func ValidTeacherIntroVideoStatus(value string) bool {
	switch TeacherIntroVideoStatus(value) {
	case TeacherIntroVideoStatusSubmitted, TeacherIntroVideoStatusApproved, TeacherIntroVideoStatusRejected, TeacherIntroVideoStatusDeleted:
		return true
	default:
		return false
	}
}

const MaxIntroVideoDurationSeconds = 60
const MaxIntroVideoBytes = 20 << 20
