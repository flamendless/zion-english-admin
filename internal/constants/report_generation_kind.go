package constants

type ReportGenerationKind string

const (
	ReportGenerationKindTeacher         ReportGenerationKind = "teacher"
	ReportGenerationKindSummary         ReportGenerationKind = "summary"
	ReportGenerationKindTeachersPayroll ReportGenerationKind = "teachers_payroll"
)

const ReportHistorySummaryLabel = "Payroll summary"

const ReportHistoryTeachersPayrollLabel = "Teachers payroll"
