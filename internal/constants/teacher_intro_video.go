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

type TeacherIntroVideoSourceType string

const (
	TeacherIntroVideoSourceUpload     TeacherIntroVideoSourceType = "upload"
	TeacherIntroVideoSourceGoogleDrive TeacherIntroVideoSourceType = "google_drive"
	TeacherIntroVideoSourceYouTube     TeacherIntroVideoSourceType = "youtube"
)

var TeacherIntroVideoSourceTypes = []TeacherIntroVideoSourceType{
	TeacherIntroVideoSourceUpload,
	TeacherIntroVideoSourceGoogleDrive,
	TeacherIntroVideoSourceYouTube,
}

func (s TeacherIntroVideoSourceType) Label() string {
	switch s {
	case TeacherIntroVideoSourceUpload:
		return "Upload"
	case TeacherIntroVideoSourceGoogleDrive:
		return "Google Drive"
	case TeacherIntroVideoSourceYouTube:
		return "YouTube"
	default:
		return string(s)
	}
}

func ValidTeacherIntroVideoSourceType(value string) bool {
	switch TeacherIntroVideoSourceType(value) {
	case TeacherIntroVideoSourceUpload, TeacherIntroVideoSourceGoogleDrive, TeacherIntroVideoSourceYouTube:
		return true
	default:
		return false
	}
}

const MaxIntroVideoDurationSeconds = 60
const MaxIntroVideoBytes = 20 << 20
