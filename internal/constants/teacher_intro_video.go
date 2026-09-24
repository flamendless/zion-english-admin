package constants

import "fmt"

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

const MaxIntroVideoDurationSeconds = 120
const MaxIntroVideoBytes = 200 << 20

const IntroVideoStoredExt = ".mp4"
const IntroVideoStoredMIME = "video/mp4"

func MaxIntroVideoSizeMB() int {
	return int(MaxIntroVideoBytes / (1 << 20))
}

func IntroVideoFileTooLargeMessage() string {
	return fmt.Sprintf("[INTRO VIDEO] File is too large. Maximum size is %d MB.", MaxIntroVideoSizeMB())
}

func MaxIntroVideoDurationLabel() string {
	minutes := MaxIntroVideoDurationSeconds / 60
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}

func IntroVideoTooLongMessage() string {
	return fmt.Sprintf("[INTRO VIDEO] Video is too long. Maximum duration is %s.", MaxIntroVideoDurationLabel())
}
