package frontend

import "zion-english/internal/constants"

var SummaryReportKindOptions = []StatusOption{
	{Value: string(constants.SummaryReportKindParents), Label: "Parents summary"},
	{Value: string(constants.SummaryReportKindTeachersPayroll), Label: "Teachers payroll"},
}
