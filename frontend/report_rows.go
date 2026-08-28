package frontend

import (
	"fmt"
	"strings"

	"zion-english/internal/utils"
)

const downloadDisabledTooltip = "Generate a report first to enable download"

type ReportEarningData struct {
	Currency string
	Total    float64
}

type ReportRowData struct {
	TeacherID        string
	TeacherName      string
	TeacherAvatar    AvatarProps
	ConductedClasses int64
	TotalClasses     int64
	Earnings         []ReportEarningData
	DownloadReady    bool
	Filename         string
}

func (r ReportRowData) ClassesLabel() string {
	return fmt.Sprintf("%d/%d", r.ConductedClasses, r.TotalClasses)
}

func (r ReportRowData) GenerateLabel() string {
	if r.DownloadReady {
		return "Regenerate"
	}
	return "Generate report"
}

func (r ReportRowData) GenerateKind() IconKind {
	if r.DownloadReady {
		return IconKindRegenerate
	}
	return IconKindGenerate
}

func (r ReportRowData) EarningsHTML() string {
	if len(r.Earnings) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(r.Earnings))
	for _, item := range r.Earnings {
		parts = append(parts, utils.FormatCurrency(item.Total, item.Currency))
	}
	return strings.Join(parts, "<br>")
}

func (r ReportRowData) DownloadURL() string {
	if !r.DownloadReady || r.Filename == "" {
		return "#"
	}
	return utils.URL("/download/processed?filename=" + r.Filename)
}
