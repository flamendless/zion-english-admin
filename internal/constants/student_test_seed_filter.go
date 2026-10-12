package constants

type StudentTestSeedFilter string

const (
	StudentTestSeedFilterExclude StudentTestSeedFilter = ""
	StudentTestSeedFilterOnly    StudentTestSeedFilter = "only"
	StudentTestSeedFilterAll     StudentTestSeedFilter = "all"
)

func ValidStudentTestSeedFilter(value string) bool {
	switch StudentTestSeedFilter(value) {
	case StudentTestSeedFilterExclude, StudentTestSeedFilterOnly, StudentTestSeedFilterAll:
		return true
	default:
		return false
	}
}

func (f StudentTestSeedFilter) Label() string {
	switch f {
	case StudentTestSeedFilterExclude:
		return "Exclude test/seed"
	case StudentTestSeedFilterOnly:
		return "Test/seed only"
	case StudentTestSeedFilterAll:
		return "All students"
	default:
		return ""
	}
}
