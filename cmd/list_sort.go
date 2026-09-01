package cmd

import (
	"database/sql"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

func sortSQLKey(sort utils.SortParams) string {
	return utils.SortKey(sort.By, sort.Order)
}

func paginateSlice[T any](items []T, page utils.Page) []T {
	start := page.Offset()
	if start >= len(items) {
		return nil
	}
	end := start + page.Size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func sortStudentRows(rows []queries.GetStudentsFilteredRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetStudentsFilteredRow) int {
		switch sort.By {
		case "name":
			return utils.CompareStrings(a.Name, b.Name)
		case "status":
			return utils.CompareStrings(a.Status, b.Status)
		case "rate_per_class":
			return utils.CompareFloat64(a.RatePerClass, b.RatePerClass)
		case "created_at":
			return utils.CompareStrings(nullTimeValue(a.CreatedAt), nullTimeValue(b.CreatedAt))
		default:
			return utils.CompareStrings(nullTimeValue(a.CreatedAt), nullTimeValue(b.CreatedAt))
		}
	})
}

func sortTeacherRows(rows []queries.GetTeachersFilteredRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetTeachersFilteredRow) int {
		switch sort.By {
		case "name":
			return utils.CompareStrings(teacherFullName(a.FirstName, a.MiddleName, a.LastName), teacherFullName(b.FirstName, b.MiddleName, b.LastName))
		case "email":
			return utils.CompareStrings(a.Email, b.Email)
		case "status":
			return utils.CompareStrings(teacherSortStatus(a), teacherSortStatus(b))
		case "created_at":
			return utils.CompareStrings(nullTimeValue(a.CreatedAt), nullTimeValue(b.CreatedAt))
		default:
			return utils.CompareStrings(teacherFullName(a.FirstName, a.MiddleName, a.LastName), teacherFullName(b.FirstName, b.MiddleName, b.LastName))
		}
	})
}

func teacherSortStatus(row queries.GetTeachersFilteredRow) string {
	if row.Deleted != 0 {
		return "deleted"
	}
	return row.Status
}

func sortMyStudentRows(rows []queries.GetStudentsByTeacherIDFilteredRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetStudentsByTeacherIDFilteredRow) int {
		switch sort.By {
		case "name":
			return utils.CompareStrings(a.Name, b.Name)
		case "contact":
			return utils.CompareStrings(a.Contact.String, b.Contact.String)
		case "status":
			return utils.CompareStrings(a.Status, b.Status)
		case "rate_per_class":
			return utils.CompareFloat64(a.RatePerClass, b.RatePerClass)
		default:
			return utils.CompareStrings(a.Name, b.Name)
		}
	})
}

func sortClassListRows(rows []queries.GetClassesListFilteredRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetClassesListFilteredRow) int {
		switch sort.By {
		case "date":
			if c := utils.CompareStrings(a.Date, b.Date); c != 0 {
				return c
			}
			return utils.CompareStrings(a.StartTime.String, b.StartTime.String)
		case "start_time":
			if c := utils.CompareStrings(a.StartTime.String, b.StartTime.String); c != 0 {
				return c
			}
			return utils.CompareStrings(a.Date, b.Date)
		case "student_name":
			return utils.CompareStrings(a.StudentName, b.StudentName)
		case "teacher_name":
			return utils.CompareStrings(a.TeacherName, b.TeacherName)
		case "rate":
			return utils.CompareFloat64(a.Rate, b.Rate)
		case "status":
			return utils.CompareStrings(a.Status, b.Status)
		default:
			if c := utils.CompareStrings(a.Date, b.Date); c != 0 {
				return c
			}
			return utils.CompareStrings(a.StartTime.String, b.StartTime.String)
		}
	})
}

func sortDocumentRows(rows []queries.GetAllTeacherDocumentsFilteredRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetAllTeacherDocumentsFilteredRow) int {
		switch sort.By {
		case "filename":
			return utils.CompareStrings(a.OriginalFilename, b.OriginalFilename)
		case "type":
			return utils.CompareStrings(a.Type, b.Type)
		case "file_size":
			return utils.CompareInt64(a.FileSize, b.FileSize)
		case "status":
			return utils.CompareStrings(a.Status, b.Status)
		case "uploaded_at":
			return utils.CompareStrings(nullTimeValue(a.UploadedAt), nullTimeValue(b.UploadedAt))
		default:
			return utils.CompareStrings(nullTimeValue(a.UploadedAt), nullTimeValue(b.UploadedAt))
		}
	})
}

func sortTeacherDocumentRows(rows []queries.TblTeacherDocument, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.TblTeacherDocument) int {
		switch sort.By {
		case "filename":
			return utils.CompareStrings(a.OriginalFilename, b.OriginalFilename)
		case "type":
			return utils.CompareStrings(a.Type, b.Type)
		case "file_size":
			return utils.CompareInt64(a.FileSize, b.FileSize)
		case "status":
			return utils.CompareStrings(a.Status, b.Status)
		case "uploaded_at":
			return utils.CompareStrings(nullTimeValue(a.UploadedAt), nullTimeValue(b.UploadedAt))
		default:
			return utils.CompareStrings(nullTimeValue(a.UploadedAt), nullTimeValue(b.UploadedAt))
		}
	})
}

func sortProcessingLogRows(rows []queries.TblProcessingLog, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.TblProcessingLog) int {
		switch sort.By {
		case "id":
			return utils.CompareInt64(a.ID, b.ID)
		case "name":
			return utils.CompareStrings(a.Name, b.Name)
		case "created_at":
			return utils.CompareStrings(nullTimeValue(a.CreatedAt), nullTimeValue(b.CreatedAt))
		default:
			return utils.CompareStrings(nullTimeValue(a.CreatedAt), nullTimeValue(b.CreatedAt))
		}
	})
}

func sortSystemLogRows(rows []queries.GetAllLogsFilteredRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetAllLogsFilteredRow) int {
		switch sort.By {
		case "id":
			return utils.CompareInt64(a.ID, b.ID)
		case "module":
			return utils.CompareStrings(a.Module, b.Module)
		case "created_at":
			return utils.CompareStrings(a.CreatedAt, b.CreatedAt)
		default:
			return utils.CompareStrings(a.CreatedAt, b.CreatedAt)
		}
	})
}

func sortSystemLogByUserRows(rows []queries.GetLogsByCreatedByFilteredRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetLogsByCreatedByFilteredRow) int {
		switch sort.By {
		case "id":
			return utils.CompareInt64(a.ID, b.ID)
		case "module":
			return utils.CompareStrings(a.Module, b.Module)
		case "created_at":
			return utils.CompareStrings(a.CreatedAt, b.CreatedAt)
		default:
			return utils.CompareStrings(a.CreatedAt, b.CreatedAt)
		}
	})
}

func sortNotificationRows(rows []queries.TblNotification, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.TblNotification) int {
		switch sort.By {
		case "from_name":
			return utils.CompareStrings(a.FromName, b.FromName)
		case "message":
			return utils.CompareStrings(a.Message, b.Message)
		case "read":
			return utils.CompareInt64(a.Read, b.Read)
		case "created_at":
			if c := utils.CompareStrings(a.CreatedAt, b.CreatedAt); c != 0 {
				return c
			}
			return utils.CompareInt64(a.ID, b.ID)
		default:
			if c := utils.CompareStrings(a.CreatedAt, b.CreatedAt); c != 0 {
				return c
			}
			return utils.CompareInt64(a.ID, b.ID)
		}
	})
}

func sortAnnouncementRows(rows []queries.GetAnnouncementsPagedRow, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b queries.GetAnnouncementsPagedRow) int {
		switch sort.By {
		case "title":
			return utils.CompareStrings(a.Title, b.Title)
		case "level":
			return utils.CompareStrings(a.Level, b.Level)
		case "start_date":
			if c := utils.CompareStrings(a.StartDate, b.StartDate); c != 0 {
				return c
			}
			return utils.CompareInt64(a.ID, b.ID)
		case "status":
			return utils.CompareStrings(a.Status, b.Status)
		default:
			if c := utils.CompareStrings(a.StartDate, b.StartDate); c != 0 {
				return c
			}
			return utils.CompareInt64(a.ID, b.ID)
		}
	})
}

func sortReportRows(rows []frontend.ReportRowData, sort utils.SortParams) {
	utils.SortSlice(rows, sort.Order, func(a, b frontend.ReportRowData) int {
		switch sort.By {
		case "total_classes":
			if c := utils.CompareInt64(a.TotalClasses, b.TotalClasses); c != 0 {
				return c
			}
			return utils.CompareStrings(a.TeacherName, b.TeacherName)
		case "earnings":
			if c := utils.CompareFloat64(reportPrimaryEarning(a.Earnings), reportPrimaryEarning(b.Earnings)); c != 0 {
				return c
			}
			return utils.CompareStrings(a.TeacherName, b.TeacherName)
		case "teacher_name":
			return utils.CompareStrings(a.TeacherName, b.TeacherName)
		default:
			return utils.CompareStrings(a.TeacherName, b.TeacherName)
		}
	})
}

func reportPrimaryEarning(earnings []frontend.ReportEarningData) float64 {
	if len(earnings) == 0 {
		return 0
	}
	return earnings[0].Total
}

func teacherFullName(first, middle, last string) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(first) != "" {
		parts = append(parts, strings.TrimSpace(first))
	}
	if strings.TrimSpace(middle) != "" {
		parts = append(parts, strings.TrimSpace(middle))
	}
	if strings.TrimSpace(last) != "" {
		parts = append(parts, strings.TrimSpace(last))
	}
	return strings.Join(parts, " ")
}

func nullStringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func nullTimeValue(v sql.NullTime) string {
	if v.Valid {
		return v.Time.UTC().Format("2006-01-02 15:04:05")
	}
	return ""
}
