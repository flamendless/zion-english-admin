package frontend

import (
	"strconv"
	"strings"

	"zion-english/internal/constants"
)

type PayrollSessionListItem struct {
	ID            string
	PeriodLabel   string
	Status        constants.PayrollSessionStatus
	RolesLabel    string
	TeacherCount  int64
	ProgressLabel string
	StartedAt     string
	FinishedAt    string
	StartedByName string
	IsActive      bool
}

type PayrollListData struct {
	Sessions           []PayrollSessionListItem
	Pagination         PaginationData
	PrevURL            string
	NextURL            string
	HasActiveSession   bool
	ActiveSessionID    string
	ActiveSessionLabel string
}

type PayrollSessionData struct {
	SessionID           string
	PeriodLabel         string
	StartDate           string
	EndDate             string
	Status              constants.PayrollSessionStatus
	RoleTeacher         bool
	RoleAdmin           bool
	IsActive            bool
	CanFinish           bool
	FinishBlockedReason string
	Teachers            []PayrollTeacherRow
}

const payrollReportActionsDisabledTooltip = "Confirm daily logs on the teacher dashboard first"

type PayrollTeacherRow struct {
	TeacherID          string
	TeacherName        string
	Avatar             AvatarProps
	Status             constants.PayrollTeacherStatus
	TotalClasses       int64
	EarningsHTML       string
	SessionID          string
	StartDate          string
	EndDate            string
	IsActiveSession    bool
	ShowReportActions  bool
	CanViewReport      bool
	CanGenerate        bool
	CanAudit           bool
	CanSendSalary      bool
	DownloadReady      bool
	DownloadURL        string
	GenerateLabel      string
	GenerateKind       IconKind
	SendPaymentTooltip string
}

func (r *PayrollTeacherRow) ApplyActionGates(report ReportRowData) {
	step := r.Status.Step()
	r.ShowReportActions = step >= 1
	r.CanViewReport = step >= 2
	r.CanGenerate = step >= 2
	r.GenerateLabel = report.GenerateLabel()
	r.GenerateKind = report.GenerateKind()
	r.DownloadReady = report.DownloadReady && r.CanGenerate
	r.DownloadURL = report.DownloadURL()
	r.CanAudit = r.IsActiveSession && r.Status == constants.PayrollTeacherStatusDoneDailyLogs
	r.CanSendSalary = r.IsActiveSession && r.Status == constants.PayrollTeacherStatusAudited && report.CanSendPayment()
	r.SendPaymentTooltip = report.SendPaymentDisabledTooltip
}

type PayrollDailyLogsData struct {
	SessionID   string
	PeriodLabel string
}

type PayrollTeacherStatusEventItem struct {
	ChangedAtLabel string
	FromLabel      string
	ToLabel        string
	ActionLabel    string
	ChangedByName  string
}

type PayrollTeacherEditModalData struct {
	SessionID     string
	TeacherID     string
	TeacherName   string
	CurrentStatus constants.PayrollTeacherStatus
	Events        []PayrollTeacherStatusEventItem
}

func ReportRolesFilterSummary(roleTeacher, roleAdmin bool) string {
	if !roleTeacher && !roleAdmin {
		return "All"
	}
	parts := make([]string, 0, 2)
	if roleTeacher {
		parts = append(parts, featureFlagRoleLabel(constants.TeacherRoleTeacher))
	}
	if roleAdmin {
		parts = append(parts, featureFlagRoleLabel(constants.TeacherRoleAdmin))
	}
	return strings.Join(parts, ", ")
}

func (r PayrollTeacherRow) ClassesLabel() string {
	return strconv.FormatInt(r.TotalClasses, 10)
}
