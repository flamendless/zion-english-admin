package frontend

import (
	"fmt"

	"zion-english/internal/models"
	"zion-english/internal/utils"
)

type ClassRecordRowData struct {
	ID              int64
	StudentID       int64
	TeacherID       int64
	StudentName     string
	TeacherName     string
	Date            string
	StartTime       string
	EndTime         string
	DurationMinutes int64
	Rate            float64
	Currency        string
	Status          string
	Source          string
}

func ClassRecordRowFromView(v models.ClassRecordView) ClassRecordRowData {
	return ClassRecordRowData{
		ID:              v.ID,
		StudentID:       v.StudentID,
		TeacherID:       v.TeacherID,
		StudentName:     v.StudentName,
		TeacherName:     v.TeacherName,
		Date:            v.Date,
		StartTime:       v.StartTime,
		EndTime:         v.EndTime,
		DurationMinutes: v.DurationMinutes,
		Rate:            v.Rate,
		Currency:        v.Currency,
		Status:          v.Status,
		Source:          v.Source,
	}
}

func ClassRecordRowFromViews(views []models.ClassRecordView) []ClassRecordRowData {
	rows := make([]ClassRecordRowData, 0, len(views))
	for _, v := range views {
		rows = append(rows, ClassRecordRowFromView(v))
	}
	return rows
}

func (r ClassRecordRowData) IsScheduled() bool {
	return r.Source == "scheduled"
}

func (r ClassRecordRowData) RowClass() string {
	switch {
	case r.Status == "cancelled":
		return "row-cancelled"
	case r.Status == "rescheduled":
		return "row-rescheduled"
	case r.IsScheduled():
		return "row-scheduled"
	default:
		return ""
	}
}

func (r ClassRecordRowData) RateLabel() string {
	return fmt.Sprintf("%g %s", r.Rate, r.Currency)
}

func (r ClassRecordRowData) StartTimeDisplay() string {
	if r.StartTime == "" {
		return "-"
	}
	return r.StartTime
}

func (r ClassRecordRowData) EndTimeDisplay() string {
	if r.EndTime == "" {
		return "-"
	}
	return r.EndTime
}

func (r ClassRecordRowData) ScheduledItemData() ScheduledClassItemData {
	item := ScheduledClassItemData{
		ID:              r.ID,
		StudentID:       r.StudentID,
		TeacherID:       r.TeacherID,
		StudentName:     r.StudentName,
		TeacherName:     r.TeacherName,
		ScheduledDate:   r.Date,
		StartTime:       r.StartTime,
		EndTime:         r.EndTime,
		DurationMinutes: r.DurationMinutes,
		Rate:            r.Rate,
		Currency:        r.Currency,
		DeleteFrom:      "classes",
	}
	item.TimeRange = FormatScheduledClassTimeRange(item.StartTime, item.EndTime, item.DurationMinutes)
	return item
}

func (r ClassRecordRowData) ViewURL() string {
	if r.IsScheduled() {
		return utils.URL("/schedule/" + fmt.Sprintf("%d", r.ID) + "/edit")
	}
	return utils.URL("/classes/" + fmt.Sprintf("%d", r.ID) + "/view")
}

func (r ClassRecordRowData) EditURL() string {
	if r.IsScheduled() {
		return utils.URL("/schedule/" + fmt.Sprintf("%d", r.ID) + "/edit")
	}
	return utils.URL("/classes/" + fmt.Sprintf("%d", r.ID) + "/edit")
}

func (r ClassRecordRowData) DeleteURL() string {
	if r.IsScheduled() {
		return ScheduledClassDeleteURL(r.ID)
	}
	return ClassRecordDeleteURL(r.ID)
}

type PaginationData struct {
	HasPrev    bool
	PrevPage   int
	HasNext    bool
	NextPage   int
	Number     int
	TotalPages int
	Total      int64
}

func BuildPaginationData(pageNum, pageSize int, total int64) PaginationData {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}
	return PaginationData{
		Number:     pageNum,
		TotalPages: totalPages,
		Total:      total,
		HasPrev:    pageNum > 1,
		HasNext:    pageNum < totalPages,
		PrevPage:   pageNum - 1,
		NextPage:   pageNum + 1,
	}
}
