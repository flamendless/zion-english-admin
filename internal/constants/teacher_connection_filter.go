package constants

type TeacherConnectionFilter string

const (
	TeacherConnectionFilterMissingZoom   TeacherConnectionFilter = "missing_zoom"
	TeacherConnectionFilterMissingGoogle TeacherConnectionFilter = "missing_google"
	TeacherConnectionFilterMissingBoth   TeacherConnectionFilter = "missing_both"
)

func ValidTeacherConnectionFilter(value string) bool {
	switch TeacherConnectionFilter(value) {
	case TeacherConnectionFilterMissingZoom, TeacherConnectionFilterMissingGoogle, TeacherConnectionFilterMissingBoth:
		return true
	default:
		return false
	}
}
