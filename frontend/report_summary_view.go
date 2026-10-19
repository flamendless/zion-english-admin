package frontend

import (
	"fmt"
	"strings"

	"zion-english/internal/constants"
)

type ParentsSummaryClassRow struct {
	Status      constants.ClassStatus
	Date        string
	Duration    string
	ParentRate  string
	ClassRate   string
}

type ParentsSummaryStudentBlock struct {
	StudentName  string
	Classes      []ParentsSummaryClassRow
	TotalLabel   string
	TotalDisplay string
	HasTotal     bool
}

type ParentsSummaryTeacherBlock struct {
	TeacherName string
	Avatar      AvatarProps
	Students    []ParentsSummaryStudentBlock
}

type TeachersPayrollSummaryRow struct {
	TeacherName     string
	TeacherAvatar   AvatarProps
	CutoffLabel     string
	TotalClasses    int64
	EarningsDisplay string
}

func (r TeachersPayrollSummaryRow) ClassesLabel() string {
	return fmt.Sprintf("%d", r.TotalClasses)
}

type ReportSummaryViewData struct {
	Kind          constants.SummaryReportKind
	Title         string
	CutoffLabel   string
	Parents       []ParentsSummaryTeacherBlock
	Payroll       []TeachersPayrollSummaryRow
}

func (d ReportSummaryViewData) IsParents() bool {
	return d.Kind == constants.SummaryReportKindParents
}

func (d ReportSummaryViewData) IsPayroll() bool {
	return d.Kind == constants.SummaryReportKindTeachersPayroll
}

func summaryClassRowClass(status constants.ClassStatus) string {
	switch status {
	case constants.ClassStatusConducted:
		return "summary-class-row--conducted"
	case constants.ClassStatusCancelled:
		return "summary-class-row--cancelled"
	case constants.ClassStatusRescheduled:
		return "summary-class-row--rescheduled"
	default:
		return ""
	}
}

func (r TeachersPayrollSummaryRow) EarningsHTML() string {
	if strings.TrimSpace(r.EarningsDisplay) == "" || r.EarningsDisplay == constants.DisplayNotSet {
		return constants.DisplayNotSet
	}
	return strings.ReplaceAll(r.EarningsDisplay, ", ", "<br>")
}
