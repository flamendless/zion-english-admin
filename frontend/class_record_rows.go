package frontend

import (
	"fmt"

	"zion-english/internal/constants"
	"zion-english/internal/models"
	"zion-english/internal/utils"
)

type ClassRecordRowData struct {
	ID              int64
	StudentID       int64
	TeacherID       int64
	StudentName     string
	StudentNumber   string
	StudentAvatar   AvatarProps
	TeacherName     string
	TeacherAvatar   AvatarProps
	Date            string
	StartTime       string
	EndTime         string
	DurationMinutes int64
	Rate            float64
	Currency        string
	Status          constants.ClassListFilterStatus
	Overdue         bool
	Reason          string
	Notes           string
	Source          ClassRecordSource
}

func ClassRecordRowFromView(v models.ClassRecordView) ClassRecordRowData {
	return ClassRecordRowData{
		ID:            v.ID,
		StudentID:     v.StudentID,
		TeacherID:     v.TeacherID,
		StudentName:   v.StudentName,
		StudentNumber: v.StudentNumber,
		StudentAvatar: AvatarProps{
			Size:          "sm",
			Initials:      v.StudentAvatar.Initials,
			AssignedColor: v.StudentAvatar.AssignedColor,
			PictureURL:    v.StudentAvatar.PictureURL,
			HasPicture:    v.StudentAvatar.HasPicture,
			Alt:           v.StudentAvatar.Alt,
			RoleBadge:     v.StudentAvatar.RoleBadge,
		},
		TeacherName: v.TeacherName,
		TeacherAvatar: AvatarProps{
			Size:                 "sm",
			Initials:             v.TeacherAvatar.Initials,
			AssignedColor:        v.TeacherAvatar.AssignedColor,
			PictureURL:           v.TeacherAvatar.PictureURL,
			HasPicture:           v.TeacherAvatar.HasPicture,
			Alt:                  v.TeacherAvatar.Alt,
			RoleBadge:            v.TeacherAvatar.RoleBadge,
			TeacherAccountStatus: constants.TeacherStatus(v.TeacherAvatar.TeacherAccountStatus),
		},
		Date:            v.Date,
		StartTime:       v.StartTime,
		EndTime:         v.EndTime,
		DurationMinutes: v.DurationMinutes,
		Rate:            v.Rate,
		Currency:        v.Currency,
		Status:          constants.ClassListFilterStatus(v.Status),
		Reason:          v.Reason,
		Notes:           v.Notes,
		Source:          ClassRecordSource(v.Source),
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
	return r.Source == ClassRecordSourceScheduled
}

func (r ClassRecordRowData) IsDeleted() bool {
	return r.Status == constants.ClassListFilterDeleted
}

func (r ClassRecordRowData) RowClass() string {
	var classes []string
	switch {
	case r.IsDeleted():
		classes = append(classes, "row-deleted")
	case r.Status == constants.ClassListFilterCancelled:
		classes = append(classes, "row-cancelled")
	case r.Status == constants.ClassListFilterRescheduled:
		classes = append(classes, "row-rescheduled")
	case r.IsScheduled():
		classes = append(classes, "row-scheduled")
	}
	if c := rowClassForCutoffDate(r.Date); c != "" {
		classes = append(classes, c)
	}
	return joinRowClasses(classes...)
}

func (r ClassRecordRowData) RateLabel() string {
	return fmt.Sprintf("%g %s", r.Rate, r.Currency)
}

func (r ClassRecordRowData) StartTimeDisplay() string {
	if r.StartTime == "" {
		return constants.DisplayNotSet
	}
	return r.StartTime
}

func (r ClassRecordRowData) EndTimeDisplay() string {
	if r.EndTime == "" {
		return constants.DisplayNotSet
	}
	return r.EndTime
}

func (r ClassRecordRowData) TimeRangeDisplay() string {
	return FormatScheduledClassTimeRange(r.StartTime, r.EndTime, r.DurationMinutes)
}

func (r ClassRecordRowData) timeRangeEnd() string {
	end := normalizeDisplayTime(r.EndTime)
	if end == "" {
		start := normalizeDisplayTime(r.StartTime)
		if start != "" && r.DurationMinutes > 0 {
			end = utils.EndTimeFromStartAndDuration(start, r.DurationMinutes)
		}
	}
	return end
}

func (r ClassRecordRowData) TimeRangeStartLine() string {
	start := normalizeDisplayTime(r.StartTime)
	if start == "" {
		return constants.DisplayNotSet
	}
	return start + " -"
}

func (r ClassRecordRowData) TimeRangeEndLine() string {
	end := r.timeRangeEnd()
	if end == "" {
		return constants.DisplayNotSet
	}
	return end
}

func (r ClassRecordRowData) StatusIconLabel() string {
	if r.Overdue || r.Status == constants.ClassListFilterOverdue {
		return constants.ClassListFilterOverdue.Label()
	}
	return r.Status.Label()
}

func (r ClassRecordRowData) StatusIconKind() IconKind {
	if r.Overdue || r.Status == constants.ClassListFilterOverdue {
		return IconKindOverdue
	}
	switch r.Status {
	case constants.ClassListFilterConducted:
		return IconKindConduct
	case constants.ClassListFilterScheduled:
		return IconKindScheduled
	case constants.ClassListFilterCancelled:
		return IconKindCancel
	case constants.ClassListFilterRescheduled:
		return IconKindRescheduled
	case constants.ClassListFilterDeleted:
		return IconKindDelete
	default:
		return ""
	}
}

func (r ClassRecordRowData) NoteOrReasonDisplay() string {
	switch r.Status {
	case constants.ClassListFilterConducted:
		return r.Notes
	default:
		return r.Reason
	}
}

func (r ClassRecordRowData) ScheduledItemData() ScheduledClassItemData {
	item := ScheduledClassItemData{
		ID:              r.ID,
		StudentID:       r.StudentID,
		TeacherID:       r.TeacherID,
		StudentName:     utils.StudentDisplayName(r.StudentName, r.StudentNumber),
		TeacherName:     r.TeacherName,
		ScheduledDate:   r.Date,
		StartTime:       r.StartTime,
		EndTime:         r.EndTime,
		DurationMinutes: r.DurationMinutes,
		Rate:            r.Rate,
		Currency:        r.Currency,
		DeleteFrom:      ClassActionContextClasses,
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
