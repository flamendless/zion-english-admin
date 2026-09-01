package frontend

import (
	"fmt"
	"strings"

	"zion-english/internal/utils"
)

type ReportAllTeachersRow struct {
	TeacherName      string
	TeacherAvatar    AvatarProps
	ConductedClasses int64
	TotalClasses     int64
	Earnings         []ReportEarningData
}

type ReportAllTeachersData struct {
	CutoffLabel string
	Rows        []ReportAllTeachersRow
	EmptyMsg    string
}

func (r ReportAllTeachersRow) ClassesLabel() string {
	return fmt.Sprintf("%d/%d", r.ConductedClasses, r.TotalClasses)
}

func (r ReportAllTeachersRow) EarningsHTML() string {
	if len(r.Earnings) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(r.Earnings))
	for _, item := range r.Earnings {
		parts = append(parts, utils.FormatCurrency(item.Total, item.Currency))
	}
	return strings.Join(parts, "<br>")
}
