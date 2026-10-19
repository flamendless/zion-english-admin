package constants

type SummaryReportKind string

const (
	SummaryReportKindParents         SummaryReportKind = "parents_summary"
	SummaryReportKindTeachersPayroll SummaryReportKind = "teachers_payroll"
)

func ParseSummaryReportKind(value string) SummaryReportKind {
	switch SummaryReportKind(value) {
	case SummaryReportKindParents, SummaryReportKindTeachersPayroll:
		return SummaryReportKind(value)
	default:
		return SummaryReportKindParents
	}
}
