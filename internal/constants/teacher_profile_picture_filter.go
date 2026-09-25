package constants

type TeacherProfilePictureFilter string

const (
	TeacherProfilePictureFilterMissing TeacherProfilePictureFilter = "missing"
)

func ValidTeacherProfilePictureFilter(value string) bool {
	switch TeacherProfilePictureFilter(value) {
	case TeacherProfilePictureFilterMissing:
		return true
	default:
		return false
	}
}
