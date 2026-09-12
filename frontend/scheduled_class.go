package frontend

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"zion-english/internal/constants"
	"zion-english/internal/models"
	"zion-english/internal/scheduledclass"
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
	RoomURL           string
	RoomPasscode      string
	CalendarEventURL  string
	TimeRange         string
	Overdue         bool
	ShowZoomWarning bool
	DeleteFrom      ClassActionContext
	SeriesID        int64
	SeriesTotalCount int64
	SeriesFutureCount int64
}

func ScheduledClassItemFromView(v models.ScheduledClassView, gracePeriodMinutes int64) ScheduledClassItemData {
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
		RoomURL:          v.RoomURL,
		RoomPasscode:     v.RoomPasscode,
		CalendarEventURL: v.CalendarEventURL,
		TimeRange:       FormatScheduledClassTimeRange(v.StartTime, v.EndTime, v.DurationMinutes),
		ShowZoomWarning: v.RoomURL == "" && v.DurationMinutes > ZoomMaxAutoMinutes,
		DeleteFrom:      ClassActionContextSchedule,
		SeriesID:        v.SeriesID,
	}
	item.Overdue = IsScheduledClassOverdue(item, gracePeriodMinutes)
	return item
}

func TimelineHourLabels() []string {
	labels := make([]string, 0, TimelineDayEndHour-TimelineDayStartHour+1)
	for hour := TimelineDayStartHour; hour <= TimelineDayEndHour; hour++ {
		labels = append(labels, formatTimelineHourLabel(hour))
	}
	return labels
}

func formatTimelineHourLabel(hour int) string {
	if hour == 0 || hour == 24 {
		return "12 AM"
	}
	if hour < 12 {
		return fmt.Sprintf("%d AM", hour)
	}
	if hour == 12 {
		return "12 PM"
	}
	return fmt.Sprintf("%d PM", hour%12)
}

func timelineBarPosition(startTime string, durationMinutes int64) (leftPct, widthPct float64, ok bool) {
	startMins, err := utils.MinutesSinceMidnight(startTime)
	if err != nil {
		return 0, 0, false
	}
	dayStart := int64(TimelineDayStartHour * 60)
	dayEnd := int64(TimelineDayEndHour * 60)
	span := float64(dayEnd - dayStart)
	if span <= 0 {
		return 0, 0, false
	}

	endMins := startMins + durationMinutes
	if durationMinutes <= 0 {
		endMins = startMins + 60
	}

	barStart := startMins
	if barStart < dayStart {
		barStart = dayStart
	}
	barEnd := endMins
	if barEnd > dayEnd {
		barEnd = dayEnd
	}
	if barEnd <= barStart {
		return 0, 0, false
	}

	leftPct = float64(barStart-dayStart) / span * 100
	widthPct = float64(barEnd-barStart) / span * 100
	return leftPct, widthPct, true
}

func BuildScheduledClassDayTimeline(items []ScheduledClassItemData, emptyMessage string) ScheduledClassDayTimelineData {
	rowsByTeacher := make(map[int64]*ScheduledClassTeacherTimelineRow)
	teacherOrder := make([]int64, 0)
	outsideWindow := make([]ScheduledClassItemData, 0)

	for _, item := range items {
		row, exists := rowsByTeacher[item.TeacherID]
		if !exists {
			row = &ScheduledClassTeacherTimelineRow{
				TeacherID:     item.TeacherID,
				TeacherName:   item.TeacherName,
				TeacherAvatar: item.TeacherAvatar,
				Bars:          []ScheduledClassTimelineBarData{},
			}
			rowsByTeacher[item.TeacherID] = row
			teacherOrder = append(teacherOrder, item.TeacherID)
		}

		leftPct, widthPct, ok := timelineBarPosition(item.StartTime, item.DurationMinutes)
		if !ok {
			outsideWindow = append(outsideWindow, item)
			continue
		}
		row.Bars = append(row.Bars, ScheduledClassTimelineBarData{
			Item:      item,
			LeftPct:   leftPct,
			WidthPct:  widthPct,
			ShowLabel: widthPct >= 7,
		})
	}

	rows := make([]ScheduledClassTeacherTimelineRow, 0, len(teacherOrder))
	for _, teacherID := range teacherOrder {
		row := rowsByTeacher[teacherID]
		if len(row.Bars) == 0 {
			continue
		}
		sort.Slice(row.Bars, func(i, j int) bool {
			a, errA := utils.MinutesSinceMidnight(row.Bars[i].Item.StartTime)
			b, errB := utils.MinutesSinceMidnight(row.Bars[j].Item.StartTime)
			if errA != nil {
				a = 9999
			}
			if errB != nil {
				b = 9999
			}
			return a < b
		})
		rows = append(rows, *row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].TeacherName < rows[j].TeacherName
	})

	return ScheduledClassDayTimelineData{
		EmptyMessage:       emptyMessage,
		HourLabels:         TimelineHourLabels(),
		Teachers:           rows,
		OutsideWindowItems: outsideWindow,
	}
}

func ScheduledClassItemFromEditClassData(data EditClassData) ScheduledClassItemData {
	id, _ := strconv.ParseInt(data.RecordID, 10, 64)
	studentID, _ := strconv.ParseInt(data.StudentID, 10, 64)
	teacherID, _ := strconv.ParseInt(data.TeacherID, 10, 64)
	item := ScheduledClassItemData{
		ID:              id,
		StudentID:       studentID,
		TeacherID:       teacherID,
		StudentName:     data.StudentName,
		TeacherName:     data.TeacherName,
		TeacherAvatar:   data.TeacherAvatar,
		ScheduledDate:   data.Date,
		StartTime:       data.StartTime,
		EndTime:         data.EndTime,
		DurationMinutes: data.DurationMinutes,
		Rate:            data.Rate,
		Currency:        data.Currency,
		Status:          constants.ScheduledClassStatusScheduled,
		TimeRange:       data.TimeRangeLabel(),
		DeleteFrom:      data.ActionFrom,
		SeriesID:        data.SeriesID,
		SeriesFutureCount: data.SeriesFutureCount,
	}
	item.Overdue = IsScheduledClassOverdue(item, data.OverdueGracePeriodMinutes)
	return item
}

func (data EditClassData) ClassEditURL() string {
	return utils.URL("/classes/" + data.RecordID + "/edit")
}

func (data EditClassData) ClassDeleteURL() string {
	id, err := strconv.ParseInt(data.RecordID, 10, 64)
	if err != nil {
		return ""
	}
	return ClassRecordDeleteURL(id)
}

func ScheduledClassItemsFromViews(views []models.ScheduledClassView, gracePeriodMinutes int64) []ScheduledClassItemData {
	items := make([]ScheduledClassItemData, 0, len(views))
	for _, v := range views {
		items = append(items, ScheduledClassItemFromView(v, gracePeriodMinutes))
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

func IsScheduledClassOverdue(item ScheduledClassItemData, gracePeriodMinutes int64) bool {
	return scheduledclass.IsOverdue(
		item.Status,
		item.ScheduledDate,
		item.StartTime,
		item.DurationMinutes,
		time.Now(),
		gracePeriodMinutes,
	)
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

func FormatScheduledClassSeriesDates(dates []string) string {
	return utils.FormatCompactDateRanges(dates)
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
		"data-duration":       formatDurationMinutes(item.DurationMinutes),
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
	if item.SeriesID > 0 {
		attrs["data-series-id"] = strconv.FormatInt(item.SeriesID, 10)
		attrs["data-series-total-count"] = strconv.FormatInt(item.SeriesTotalCount, 10)
		attrs["data-series-future-count"] = strconv.FormatInt(item.SeriesFutureCount, 10)
	}
	return attrs
}
