package frontend

import (
	"fmt"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

const reportHistoryDownloadMissingTooltip = "Report file is no longer available"

var ReportHistoryKindFilterOptions = []StatusOption{
	{Value: string(constants.ReportGenerationKindSummary), Label: constants.ReportHistorySummaryLabel},
}

type ReportHistoryRowData struct {
	Kind          constants.ReportGenerationKind
	TeacherName   string
	TeacherAvatar AvatarProps
	PeriodLabel   string
	RecordCount   int64
	GeneratedAt   string
	DownloadReady bool
	Filename      string
}

func (r ReportHistoryRowData) IsSummary() bool {
	return r.Kind == constants.ReportGenerationKindSummary
}

func (r ReportHistoryRowData) RecordsLabel() string {
	return fmt.Sprintf("%d", r.RecordCount)
}

func (r ReportHistoryRowData) DownloadURL() string {
	if !r.DownloadReady || r.Filename == "" {
		return "#"
	}
	return utils.URL("/download/processed?filename=" + r.Filename)
}
