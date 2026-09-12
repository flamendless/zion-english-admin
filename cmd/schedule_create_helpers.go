package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"zion-english/internal/auth"
	"zion-english/internal/classrules"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/meetings"
	"zion-english/internal/models"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

type scheduledClassCreateParams struct {
	StudentID       int64
	TeacherID       int64
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
	Rate            float64
	Currency        string
	IsTrialClass    bool
	SeriesID        sql.NullInt64
}

type repeatSchedulePreviewRow struct {
	Date    string
	Ready   bool
	Message string
}

func parseScheduledClassBase(r *http.Request, user auth.User, role auth.Role) (scheduledClassCreateParams, error) {
	studentID, err := formInt64(r, "schedule_student", "student")
	if err != nil {
		return scheduledClassCreateParams{}, errors.New("student is required")
	}
	startTime := r.FormValue("start_time")
	endTime := r.FormValue("end_time")
	duration, err := utils.DurationMinutesFromRange(startTime, endTime)
	if err != nil {
		return scheduledClassCreateParams{}, friendlyTimeRangeError(err)
	}
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		return scheduledClassCreateParams{}, err
	}

	teacherID := user.ID
	if auth.HasAdminAccess(role) {
		teacherID, err = formInt64(r, "schedule_teacher", "teacher")
		if err != nil {
			return scheduledClassCreateParams{}, errors.New("teacher is required")
		}
	}

	req := models.ScheduledClassRequest{
		StudentID:       studentID,
		TeacherID:       teacherID,
		StartTime:       normalizeScheduleStartTime(startTime),
		DurationMinutes: duration,
		Rate:            rate,
		Currency:        r.FormValue("currency"),
		IsTrialClass:    formIsTrialClass(r),
	}
	models.ApplyScheduledTrialClassRate(&req)
	if err := validateScheduledClassRateCurrency(req.Rate, req.Currency); err != nil {
		return scheduledClassCreateParams{}, err
	}
	if req.StudentID == 0 {
		return scheduledClassCreateParams{}, errors.New("student is required")
	}
	if req.TeacherID == 0 {
		return scheduledClassCreateParams{}, errors.New("teacher is required")
	}
	if req.StartTime == "" {
		return scheduledClassCreateParams{}, errors.New("start time is required")
	}
	if _, err := utils.ParseTimeHM(req.StartTime); err != nil {
		return scheduledClassCreateParams{}, utils.ErrInvalidStartTime
	}
	if req.DurationMinutes <= 0 {
		return scheduledClassCreateParams{}, errors.New("duration must be greater than zero")
	}

	return scheduledClassCreateParams{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		StartTime:       req.StartTime,
		DurationMinutes: req.DurationMinutes,
		Rate:            req.Rate,
		Currency:        req.Currency,
		IsTrialClass:    req.IsTrialClass,
	}, nil
}

func parseScheduledDates(r *http.Request) ([]string, error) {
	rawDates := r.Form["scheduled_dates"]
	if len(rawDates) == 0 {
		if single := strings.TrimSpace(r.FormValue("scheduled_date")); single != "" {
			rawDates = []string{single}
		}
	}
	if len(rawDates) == 0 {
		return nil, errors.New("select at least two dates for repeating classes")
	}
	if len(rawDates) > constants.MaxRepeatScheduleDates {
		return nil, fmt.Errorf("select at most %d dates", constants.MaxRepeatScheduleDates)
	}

	seen := make(map[string]struct{}, len(rawDates))
	dates := make([]string, 0, len(rawDates))
	today := utils.TodayPHT()
	for _, raw := range rawDates {
		date := strings.TrimSpace(raw)
		if date == "" {
			continue
		}
		if _, err := utils.ParseDatePHT(date); err != nil {
			return nil, errors.New("invalid date format")
		}
		if date < today {
			return nil, fmt.Errorf("date %s cannot be in the past", date)
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		dates = append(dates, date)
	}
	if len(dates) < 2 {
		return nil, errors.New("select at least two dates for repeating classes")
	}
	sort.Strings(dates)
	return dates, nil
}

func parseConfirmedScheduledDates(r *http.Request) ([]string, error) {
	rawDates := r.Form["confirm_dates"]
	if len(rawDates) == 0 {
		return nil, errors.New("select at least one date to create")
	}
	if len(rawDates) > constants.MaxRepeatScheduleDates {
		return nil, fmt.Errorf("select at most %d dates", constants.MaxRepeatScheduleDates)
	}

	seen := make(map[string]struct{}, len(rawDates))
	dates := make([]string, 0, len(rawDates))
	today := utils.TodayPHT()
	for _, raw := range rawDates {
		date := strings.TrimSpace(raw)
		if date == "" {
			continue
		}
		if _, err := utils.ParseDatePHT(date); err != nil {
			return nil, errors.New("invalid date format")
		}
		if date < today {
			return nil, fmt.Errorf("date %s cannot be in the past", date)
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		dates = append(dates, date)
	}
	if len(dates) == 0 {
		return nil, errors.New("select at least one date to create")
	}
	sort.Strings(dates)
	return dates, nil
}

func previewRepeatScheduleRows(ctx context.Context, user auth.User, base scheduledClassCreateParams, dates []string) []repeatSchedulePreviewRow {
	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	rows := make([]repeatSchedulePreviewRow, 0, len(dates))
	for _, date := range dates {
		row := repeatSchedulePreviewRow{Date: date}
		if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
			StudentID:       base.StudentID,
			TeacherID:       base.TeacherID,
			Date:            date,
			StartTime:       base.StartTime,
			DurationMinutes: base.DurationMinutes,
		}); err != nil {
			row.Message = err.Error()
		} else {
			row.Ready = true
			row.Message = "Ready"
		}
		rows = append(rows, row)
	}
	return rows
}

func insertScheduledClassRow(ctx context.Context, q *queries.Queries, user auth.User, params scheduledClassCreateParams) (int64, error) {
	var startTime sql.NullString
	if params.StartTime != "" {
		startTime = sql.NullString{String: params.StartTime, Valid: true}
	}
	return q.InsertScheduledClass(ctx, queries.InsertScheduledClassParams{
		StudentID:       params.StudentID,
		TeacherID:       params.TeacherID,
		ScheduledDate:   params.ScheduledDate,
		StartTime:       startTime,
		DurationMinutes: params.DurationMinutes,
		Rate:            params.Rate,
		Currency:        params.Currency,
		IsTrialClass:    trialClassToInt64(params.IsTrialClass),
		SeriesID:        params.SeriesID,
		Reason:          sql.NullString{},
		CreatedByRole:   string(user.Role),
	})
}

func runScheduledClassSideEffects(ctx context.Context, user auth.User, scheduleID int64, params scheduledClassCreateParams, lmIDs []int64, action string) error {
	if err := saveScheduledClassLearningMaterials(ctx, user, scheduleID, lmIDs); err != nil {
		return err
	}
	if meetingSvc != nil && meetings.SupportsAutoRoom(params.DurationMinutes) {
		student, studentErr := dbRO.GetQueries().GetStudentByID(ctx, params.StudentID)
		studentName := "student"
		if studentErr == nil {
			studentName = student.Name
		}
		if err := meetingSvc.SyncRoomForSchedule(ctx, meetings.ScheduledClassMeetingInput{
			ScheduleID:      scheduleID,
			TeacherID:       params.TeacherID,
			StudentName:     studentName,
			ScheduledDate:   params.ScheduledDate,
			StartTime:       params.StartTime,
			DurationMinutes: params.DurationMinutes,
		}); err != nil {
			logs.Log().Warn("zoom room sync failed after "+action,
				zap.Error(err),
				zap.Int64("schedule_id", scheduleID),
				zap.Int64("teacher_id", params.TeacherID),
			)
		}
	}
	if params.StartTime != "" {
		syncCalendarForSchedule(ctx, scheduleID, params.TeacherID, params.StudentID, params.ScheduledDate, params.StartTime, params.DurationMinutes, params.Rate, params.Currency, action)
	}
	return nil
}

type scheduledClassEditInput struct {
	StudentID       int64
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
	Rate            float64
	Currency        string
	IsTrialClass    bool
}

func applyScheduledClassEdit(ctx context.Context, user auth.User, target queries.GetScheduledClassesInSeriesFromDateRow, input scheduledClassEditInput, lmIDs []int64) error {
	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
		ScheduleID:      target.ID,
		StudentID:       input.StudentID,
		TeacherID:       target.TeacherID,
		Date:            input.ScheduledDate,
		StartTime:       input.StartTime,
		DurationMinutes: input.DurationMinutes,
	}); err != nil {
		return err
	}

	var startTimeNull sql.NullString
	if input.StartTime != "" {
		startTimeNull = sql.NullString{String: input.StartTime, Valid: true}
	}

	if err := dbRW.GetQueries().UpdateScheduledClassSchedule(ctx, queries.UpdateScheduledClassScheduleParams{
		StudentID:       input.StudentID,
		ScheduledDate:   input.ScheduledDate,
		StartTime:       startTimeNull,
		DurationMinutes: input.DurationMinutes,
		ID:              target.ID,
	}); err != nil {
		return err
	}

	if err := dbRW.GetQueries().UpdateScheduledClassDetails(ctx, queries.UpdateScheduledClassDetailsParams{
		StudentID:    input.StudentID,
		Rate:         input.Rate,
		Currency:     input.Currency,
		IsTrialClass: trialClassToInt64(input.IsTrialClass),
		ID:           target.ID,
	}); err != nil {
		return err
	}

	if err := saveScheduledClassLearningMaterials(ctx, user, target.ID, lmIDs); err != nil {
		return err
	}

	existingStart := ""
	if target.StartTime.Valid {
		existingStart = target.StartTime.String
	}
	scheduleChanged := input.StudentID != target.StudentID ||
		input.ScheduledDate != target.ScheduledDate ||
		input.StartTime != existingStart ||
		input.DurationMinutes != target.DurationMinutes
	rateChanged := input.Rate != target.Rate || input.Currency != target.Currency

	if scheduleChanged && meetingSvc != nil && meetings.SupportsAutoRoom(input.DurationMinutes) {
		student, studentErr := dbRO.GetQueries().GetStudentByID(ctx, input.StudentID)
		studentName := "student"
		if studentErr == nil {
			studentName = student.Name
		}
		if err := meetingSvc.SyncRoomForSchedule(ctx, meetings.ScheduledClassMeetingInput{
			ScheduleID:      target.ID,
			TeacherID:       target.TeacherID,
			StudentName:     studentName,
			ScheduledDate:   input.ScheduledDate,
			StartTime:       input.StartTime,
			DurationMinutes: input.DurationMinutes,
		}); err != nil {
			logs.Log().Warn("zoom room sync failed after schedule edit",
				zap.Error(err),
				zap.Int64("schedule_id", target.ID),
				zap.Int64("teacher_id", target.TeacherID),
			)
		}
	}
	if input.StartTime != "" && (scheduleChanged || rateChanged) {
		syncCalendarForSchedule(ctx, target.ID, target.TeacherID, input.StudentID, input.ScheduledDate, input.StartTime, input.DurationMinutes, input.Rate, input.Currency, "schedule edit")
	}
	return nil
}

func cancelScheduledClassByID(ctx context.Context, user auth.User, scheduleID int64, reason, notes string) error {
	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		return errors.New("scheduled class not found")
	}
	if existing.Status != "scheduled" {
		return errors.New("only scheduled classes can be cancelled")
	}
	if meetingSvc != nil {
		_ = meetingSvc.DeleteRoomForSchedule(ctx, scheduleID, existing.TeacherID)
	}
	if calendarSvc != nil {
		_ = calendarSvc.DeleteEventForSchedule(ctx, scheduleID, existing.TeacherID)
	}
	if err := insertClassRecordFromSchedule(ctx, user, existing, scheduleID, "cancelled", reason, notes); err != nil {
		return err
	}
	insertAuditLogAs(ctx, user, "classes", fmt.Sprintf("recorded class for student id %d (teacher id %d, date %s, status cancelled)", existing.StudentID, existing.TeacherID, existing.ScheduledDate))
	insertAuditLogAs(ctx, user, "schedule", fmt.Sprintf("cancelled scheduled class id %d (student id %d, date %s)", scheduleID, existing.StudentID, existing.ScheduledDate))
	return nil
}

func deleteScheduledClassByID(ctx context.Context, user auth.User, scheduleID int64, reason string) error {
	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		return errors.New("scheduled class not found")
	}
	if existing.Status != "scheduled" {
		return errors.New("only scheduled classes can be deleted")
	}
	if meetingSvc != nil {
		_ = meetingSvc.DeleteRoomForSchedule(ctx, scheduleID, existing.TeacherID)
	}
	if calendarSvc != nil {
		_ = calendarSvc.DeleteEventForSchedule(ctx, scheduleID, existing.TeacherID)
	}
	if err := dbRW.GetQueries().SoftDeleteScheduledClass(ctx, queries.SoftDeleteScheduledClassParams{
		Reason: sql.NullString{String: reason, Valid: true},
		ID:     scheduleID,
	}); err != nil {
		return err
	}
	insertAuditLogAs(ctx, user, "schedule", fmt.Sprintf("deleted scheduled class id %d (student id %d, date %s, reason: %s)", scheduleID, existing.StudentID, existing.ScheduledDate, reason))
	return nil
}
