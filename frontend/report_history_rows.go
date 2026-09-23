package frontend

import (
	"fmt"
	"zion-english/internal/utils"
)

const reportHistoryDownloadMissingTooltip = "Report file is no longer available"

type ReportHistoryRowData struct {
	TeacherName   string
	TeacherAvatar AvatarProps
	PeriodLabel   string
	RecordCount   int64
	GeneratedAt   string
	DownloadReady bool
	Filename      string
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
