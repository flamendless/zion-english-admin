package constants

type StudentParentFilter string

const (
	StudentParentFilterMissingName StudentParentFilter = "missing_name"
	StudentParentFilterMissingRate StudentParentFilter = "missing_rate"
)

func ValidStudentParentFilter(value string) bool {
	switch StudentParentFilter(value) {
	case StudentParentFilterMissingName, StudentParentFilterMissingRate:
		return true
	default:
		return false
	}
}

func (f StudentParentFilter) Label() string {
	switch f {
	case StudentParentFilterMissingName:
		return "Missing parent name"
	case StudentParentFilterMissingRate:
		return "Missing parent rate"
	default:
		return ""
	}
}
