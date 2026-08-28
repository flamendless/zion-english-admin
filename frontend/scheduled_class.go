package frontend

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"zion-english/internal/constants"
	"zion-english/internal/models"
	"zion-english/internal/utils"
)

const (
	ZoomMaxAutoMinutes = 40
	ZoomManualWarning  = "Classes longer than 40 minutes cannot receive an automatic Zoom meeting room on a Basic account. Please create your Zoom meeting manually and share the link with your student."
)

type ScheduledClassItemData struct {
	ID              int64
	StudentID       int64
	TeacherID       int64
	StudentName     string
	TeacherName     string
	TeacherAvatar   AvatarProps
	ScheduledDate   string
	StartTime       string
	EndTime         string
	DurationMinutes int64
	Rate            float64
	Currency        string
	Status          constants.ScheduledClassStatus
	RoomURL         string
	RoomPasscode    string
	TimeRange       string
	Overdue         bool
	ShowZoomWarning bool
	DeleteFrom      ClassActionContext
}

func ScheduledClassItemFromView(v models.ScheduledClassView) ScheduledClassItemData {
	item := ScheduledClassItemData{
		ID:              v.ID,
		StudentID:       v.StudentID,
		TeacherID:       v.TeacherID,
		StudentName:     v.StudentName,
		TeacherName:     v.TeacherName,
		TeacherAvatar: AvatarProps{
			Size:          "sm",
			Initials:      v.TeacherAvatar.Initials,
			AssignedColor: v.TeacherAvatar.AssignedColor,
			PictureURL:    v.TeacherAvatar.PictureURL,
			HasPicture:    v.TeacherAvatar.HasPicture,
			Alt:           v.TeacherAvatar.Alt,
			RoleBadge:     v.TeacherAvatar.RoleBadge,
		},
		ScheduledDate:   v.ScheduledDate,
		StartTime:       v.StartTime,
		EndTime:         v.EndTime,
		DurationMinutes: v.DurationMinutes,
		Rate:            v.Rate,
		Currency:        v.Currency,
		Status:          constants.ScheduledClassStatus(v.Status),
		RoomURL:         v.RoomURL,
		RoomPasscode:    v.RoomPasscode,
		TimeRange:       FormatScheduledClassTimeRange(v.StartTime, v.EndTime, v.DurationMinutes),
		ShowZoomWarning: v.RoomURL == "" && v.DurationMinutes > ZoomMaxAutoMinutes,
		DeleteFrom:      ClassActionContextSchedule,
	}
	item.Overdue = IsScheduledClassOverdue(item)
	return item
}

func ScheduledClassItemsFromViews(views []models.ScheduledClassView) []ScheduledClassItemData {
	items := make([]ScheduledClassItemData, 0, len(views))
	for _, v := range views {
		items = append(items, ScheduledClassItemFromView(v))
	}
	sort.Slice(items, func(i, j int) bool {
		a, errA := utils.MinutesSinceMidnight(items[i].StartTime)
		b, errB := utils.MinutesSinceMidnight(items[j].StartTime)
		if errA != nil {
			a = 9999
		}
		if errB != nil {
			b = 9999
		}
		return a < b
	})
	return items
}

func FormatScheduledClassTimeRange(startTime, endTime string, durationMinutes int64) string {
	start := normalizeDisplayTime(startTime)
	end := normalizeDisplayTime(endTime)
	if end == "" && start != "" && durationMinutes > 0 {
		end = utils.EndTimeFromStartAndDuration(start, durationMinutes)
	}
	if start != "" && end != "" {
		return start + " – " + end
	}
	if start != "" {
		return start
	}
	return "Time not set"
}

func normalizeDisplayTime(value string) string {
	if value == "" {
		return ""
	}
	t, err := utils.ParseTimeHM(value)
	if err != nil {
		return ""
	}
	return t.Format(constants.TimeHMLayout)
}

func IsScheduledClassOverdue(item ScheduledClassItemData) bool {
	if item.Status != constants.ScheduledClassStatusScheduled {
		return false
	}
	if item.ScheduledDate == "" {
		return false
	}
	date, err := time.ParseInLocation(constants.DateLayout, item.ScheduledDate, constants.LocationPHT)
	if err != nil {
		return false
	}
	var endAt time.Time
	if item.StartTime != "" && item.DurationMinutes > 0 {
		startMins, err := utils.MinutesSinceMidnight(item.StartTime)
		if err != nil {
			endAt = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, constants.LocationPHT)
		} else {
			total := int(startMins) + int(item.DurationMinutes)
			endAt = time.Date(date.Year(), date.Month(), date.Day(), total/60%24, total%60, 0, 0, constants.LocationPHT)
		}
	} else {
		endAt = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, constants.LocationPHT)
	}
	return time.Now().In(constants.LocationPHT).After(endAt)
}

func FormatScheduledClassDateDisplay(date string) string {
	if date == "" {
		return "-"
	}
	t, err := time.ParseInLocation(constants.DateLayout, date, constants.LocationPHT)
	if err != nil {
		return date
	}
	return t.Format("Monday, Jan 2, 2006")
}

func ScheduledClassConductURL(id int64) string {
	return utils.URL(fmt.Sprintf("/schedule/%d/conduct", id))
}

func ScheduledClassViewURL(id int64) string {
	return utils.URL(fmt.Sprintf("/schedule/%d/view", id))
}

func ScheduledClassCancelURL(id int64) string {
	return utils.URL(fmt.Sprintf("/schedule/%d/cancel", id))
}

func ScheduledClassEditModalURL(id int64) string {
	return utils.URL(fmt.Sprintf("/schedule/%d/edit/modal", id))
}

func FormatScheduledClassRate(rate float64, currency string) string {
	return formatRate(rate, currency)
}

func scheduledClassDetailAttrs(item ScheduledClassItemData) templ.Attributes {
	attrs := templ.Attributes{
		"data-student-name":   item.StudentName,
		"data-teacher-name":   item.TeacherName,
		"data-scheduled-date": FormatScheduledClassDateDisplay(item.ScheduledDate),
		"data-time-range":     item.TimeRange,
		"data-rate":           formatRateAmount(item.Rate, item.Currency),
		"data-currency":       item.Currency,
		"data-teacher-initials":      item.TeacherAvatar.Initials,
		"data-teacher-color":         item.TeacherAvatar.AssignedColor,
		"data-teacher-picture-url":   item.TeacherAvatar.PictureURL,
		"data-teacher-has-picture":   strconv.FormatBool(item.TeacherAvatar.HasPicture),
		"data-teacher-alt":           item.TeacherAvatar.Alt,
	}
	if item.TeacherAvatar.RoleBadge != "" {
		attrs["data-teacher-role-badge"] = item.TeacherAvatar.RoleBadge
		attrs["data-teacher-role-badge-class"] = pillClass(AvatarRoleBadgeTone(item.TeacherAvatar.RoleBadge))
	}
	return attrs
}
