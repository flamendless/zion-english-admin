package frontend

import (
	"fmt"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

const reportHistoryDownloadMissingTooltip = "Report file is no longer available"

var ReportHistoryKindFilterOptions = []StatusOption{
	{Value: string(constants.ReportGenerationKindSummary), Label: constants.ReportHistorySummaryLabel},
	{Value: string(constants.ReportGenerationKindTeachersPayroll), Label: constants.ReportHistoryTeachersPayrollLabel},
}

type ReportHistoryRowData struct {
	Kind          constants.ReportGenerationKind
	TeacherName   string
	TeacherAvatar AvatarProps
	PeriodStart   string
	PeriodEnd     string
	PeriodLabel   string
	RecordCount   int64
	GeneratedAt   string
	DownloadReady bool
	Filename      string
}

func (r ReportHistoryRowData) IsSummary() bool {
	return r.IsAggregateReport()
}

func (r ReportHistoryRowData) IsAggregateReport() bool {
	switch r.Kind {
	case constants.ReportGenerationKindSummary, constants.ReportGenerationKindTeachersPayroll:
		return true
	default:
		return false
	}
}

func (r ReportHistoryRowData) RecordsLabel() string {
	return fmt.Sprintf("%d", r.RecordCount)
}

func (r ReportHistoryRowData) RowClass() string {
	return rowClassForCutoffPeriod(r.PeriodStart, r.PeriodEnd)
}

func (r ReportHistoryRowData) DownloadURL() string {
	if !r.DownloadReady || r.Filename == "" {
		return "#"
	}
	return utils.URL("/download/processed?filename=" + r.Filename)
}
