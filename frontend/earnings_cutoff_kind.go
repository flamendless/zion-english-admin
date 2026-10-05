package frontend

type EarningsCutoffKind string

const (
	EarningsCutoffKindFirst  EarningsCutoffKind = "first"
	EarningsCutoffKindSecond EarningsCutoffKind = "second"
)

func (k EarningsCutoffKind) Label() string {
	switch k {
	case EarningsCutoffKindFirst:
		return "First cutoff"
	case EarningsCutoffKindSecond:
		return "Second cutoff"
	default:
		return ""
	}
}
