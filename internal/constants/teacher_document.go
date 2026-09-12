package constants

type TeacherDocumentType string

const (
	TeacherDocumentTypeAvatar   TeacherDocumentType = "avatar"
	TeacherDocumentTypeDocument TeacherDocumentType = "document"
)

func ValidTeacherDocumentType(value string) bool {
	switch TeacherDocumentType(value) {
	case TeacherDocumentTypeAvatar, TeacherDocumentTypeDocument:
		return true
	default:
		return false
	}
}

type TeacherDocumentStatus string

const (
	TeacherDocumentStatusSubmitted TeacherDocumentStatus = "submitted"
	TeacherDocumentStatusApproved  TeacherDocumentStatus = "approved"
	TeacherDocumentStatusRejected  TeacherDocumentStatus = "rejected"
)

var TeacherDocumentStatuses = []TeacherDocumentStatus{
	TeacherDocumentStatusSubmitted,
	TeacherDocumentStatusApproved,
	TeacherDocumentStatusRejected,
}

func (s TeacherDocumentStatus) Label() string {
	switch s {
	case TeacherDocumentStatusSubmitted:
		return "Submitted"
	case TeacherDocumentStatusApproved:
		return "Approved"
	case TeacherDocumentStatusRejected:
		return "Rejected"
	default:
		return string(s)
	}
}

func ValidTeacherDocumentStatus(value string) bool {
	switch TeacherDocumentStatus(value) {
	case TeacherDocumentStatusSubmitted, TeacherDocumentStatusApproved, TeacherDocumentStatusRejected:
		return true
	default:
		return false
	}
}
