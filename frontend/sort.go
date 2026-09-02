package frontend

import "zion-english/internal/utils"

type SortOption struct {
	Value string
	Label string
}

type ListSortKind string

const (
	ListSortKindStudent       ListSortKind = "student"
	ListSortKindTeacher       ListSortKind = "teacher"
	ListSortKindMyStudent     ListSortKind = "my_student"
	ListSortKindClass         ListSortKind = "class"
	ListSortKindReport        ListSortKind = "report"
	ListSortKindDocument      ListSortKind = "document"
	ListSortKindProcessingLog ListSortKind = "processing_log"
	ListSortKindSystemLog     ListSortKind = "system_log"
	ListSortKindNotification  ListSortKind = "notification"
	ListSortKindAnnouncement  ListSortKind = "announcement"
	ListSortKindAnalytics     ListSortKind = "analytics"
)

func SortOptionsFor(kind ListSortKind) []SortOption {
	switch kind {
	case ListSortKindStudent:
		return []SortOption{
			{Value: "created_at", Label: "Date added"},
			{Value: "name", Label: "Name"},
			{Value: "status", Label: "Status"},
			{Value: "rate_per_class", Label: "Rate"},
		}
	case ListSortKindTeacher:
		return []SortOption{
			{Value: "status", Label: "Status"},
			{Value: "created_at", Label: "Date added"},
			{Value: "name", Label: "Name"},
			{Value: "email", Label: "Email"},
		}
	case ListSortKindMyStudent:
		return []SortOption{
			{Value: "name", Label: "Name"},
			{Value: "contact", Label: "Contact"},
			{Value: "rate_per_class", Label: "Rate"},
			{Value: "status", Label: "Status"},
		}
	case ListSortKindClass:
		return []SortOption{
			{Value: "date", Label: "Date"},
			{Value: "start_time", Label: "Time"},
			{Value: "student_name", Label: "Student"},
			{Value: "teacher_name", Label: "Teacher"},
			{Value: "rate", Label: "Rate"},
			{Value: "status", Label: "Status"},
		}
	case ListSortKindReport:
		return []SortOption{
			{Value: "teacher_name", Label: "Teacher"},
			{Value: "total_classes", Label: "Total classes"},
			{Value: "earnings", Label: "Earnings"},
		}
	case ListSortKindDocument:
		return []SortOption{
			{Value: "uploaded_at", Label: "Uploaded"},
			{Value: "filename", Label: "Filename"},
			{Value: "type", Label: "Type"},
			{Value: "file_size", Label: "File size"},
			{Value: "status", Label: "Status"},
		}
	case ListSortKindProcessingLog:
		return []SortOption{
			{Value: "created_at", Label: "Created at"},
			{Value: "name", Label: "Name"},
			{Value: "id", Label: "ID"},
		}
	case ListSortKindSystemLog:
		return []SortOption{
			{Value: "created_at", Label: "Created at"},
			{Value: "module", Label: "Module"},
			{Value: "id", Label: "ID"},
		}
	case ListSortKindNotification:
		return []SortOption{
			{Value: "created_at", Label: "Date"},
			{Value: "from_name", Label: "From"},
			{Value: "message", Label: "Message"},
			{Value: "read", Label: "Status"},
		}
	case ListSortKindAnnouncement:
		return []SortOption{
			{Value: "start_date", Label: "Schedule"},
			{Value: "title", Label: "Title"},
			{Value: "level", Label: "Level"},
			{Value: "status", Label: "Status"},
		}
	case ListSortKindAnalytics:
		return []SortOption{
			{Value: "name", Label: "Name"},
			{Value: "conducted", Label: "Conducted"},
			{Value: "cancelled", Label: "Cancelled"},
			{Value: "rate", Label: "Rate"},
		}
	default:
		return nil
	}
}

func DefaultSortFor(kind ListSortKind) (string, utils.SortOrder) {
	switch kind {
	case ListSortKindStudent:
		return "created_at", utils.SortOrderDesc
	case ListSortKindTeacher:
		return "status", utils.SortOrderAsc
	case ListSortKindMyStudent:
		return "name", utils.SortOrderAsc
	case ListSortKindClass:
		return "date", utils.SortOrderDesc
	case ListSortKindReport:
		return "teacher_name", utils.SortOrderAsc
	case ListSortKindDocument:
		return "uploaded_at", utils.SortOrderDesc
	case ListSortKindProcessingLog, ListSortKindSystemLog, ListSortKindNotification:
		return "created_at", utils.SortOrderDesc
	case ListSortKindAnnouncement:
		return "start_date", utils.SortOrderDesc
	case ListSortKindAnalytics:
		return "name", utils.SortOrderAsc
	default:
		return "created_at", utils.SortOrderDesc
	}
}

func sortOrderAriaLabel(order string) string {
	if order == string(utils.SortOrderAsc) {
		return "Sort ascending (A to Z)"
	}
	return "Sort descending (Z to A)"
}

func sortOrderToggleLabel(order string) string {
	if order == string(utils.SortOrderAsc) {
		return "A to Z"
	}
	return "Z to A"
}
