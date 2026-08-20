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

func ValidTeacherDocumentStatus(value string) bool {
	switch TeacherDocumentStatus(value) {
	case TeacherDocumentStatusSubmitted, TeacherDocumentStatusApproved, TeacherDocumentStatusRejected:
		return true
	default:
		return false
	}
}
