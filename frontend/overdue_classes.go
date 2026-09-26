package frontend

import (
	"fmt"
	"zion-english/internal/utils"
)

const PersistentOverdueClassesPanelCap = 8

type OverdueClassPanelItem struct {
	ID              int64
	StudentName     string
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
}

func BuildOverdueClassesPersistentPanel(
	total int64,
	rows []OverdueClassPanelItem,
	weekStart, weekEnd string,
) PersistentPanelData {
	items := make([]PersistentPanelItem, 0, len(rows)+1)
	for _, row := range rows {
		detail := FormatScheduledClassDateDisplay(row.ScheduledDate)
		timeRange := FormatScheduledClassTimeRange(row.StartTime, "", row.DurationMinutes)
		if timeRange != "" && timeRange != "Time not set" {
			detail += " · " + timeRange
		}
		items = append(items, PersistentPanelItem{
			ID:                   fmt.Sprintf("overdue-%d", row.ID),
			Label:                row.StudentName,
			Detail:               detail,
			ActionURL:            ScheduledClassViewURL(row.ID),
			ActionLabel:          "Open",
			RowStatus:            PersistentPanelRowStatusNeedsAction,
			OpenInClassViewModal: true,
		})
	}
	if total > int64(PersistentOverdueClassesPanelCap) {
		items = append(items, PersistentPanelItem{
			ID:          "overdue-view-all",
			Label:       "View all in Classes",
			Detail:      fmt.Sprintf("%d overdue classes this week", total),
			ActionURL:   utils.URL(fmt.Sprintf("/classes?status=overdue&startDate=%s&endDate=%s", weekStart, weekEnd)),
			ActionLabel: "Open",
			LinkTarget:  PersistentPanelLinkTargetSelf,
			RowStatus:   PersistentPanelRowStatusNeedsAction,
		})
	}
	return PersistentPanelData{
		StorageKey: "overdue-classes",
		Title:      "Overdue classes",
		TitleMeta:  fmt.Sprintf("(%d total)", total),
		AriaLabel:  fmt.Sprintf("Overdue classes, %d total", total),
		Items:      items,
	}
}
