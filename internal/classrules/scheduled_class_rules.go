package classrules

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

var (
	ErrDuplicateScheduled           = errors.New("a scheduled class with the same student, teacher, date, and duration already exists")
	ErrScheduleNotOwner             = errors.New("you can only manage your own scheduled classes")
	ErrTeacherScheduleConflict      = errors.New("teacher already has a class scheduled at this time")
	ErrStudentScheduleConflict      = errors.New("student already has a class scheduled at this time")
	ErrVerifyStudentAssignment      = errors.New("failed to verify student assignment")
	ErrCheckScheduledDuplicate      = errors.New("failed to check duplicate scheduled class")
	ErrCheckClassRecordDuplicate    = errors.New("failed to check duplicate class record")
	ErrCheckTeacherScheduleConflict = errors.New("failed to check teacher schedule conflicts")
	ErrCheckStudentScheduleConflict = errors.New("failed to check student schedule conflicts")
)

type ScheduleConflict struct {
	Date        string
	StartTime   string
	EndTime     string
	Status      string
	TeacherName string
	StudentName string
}

type ScheduleConflictError struct {
	Kind     error
	Conflict ScheduleConflict
}

func (e *ScheduleConflictError) Error() string {
	return fmt.Sprintf(
		"%s. Conflicting class: %s, %s - %s, status: %s, teacher: %s, student: %s",
		e.Kind.Error(),
		e.Conflict.Date,
		e.Conflict.StartTime,
		e.Conflict.EndTime,
		e.Conflict.Status,
		e.Conflict.TeacherName,
		e.Conflict.StudentName,
	)
}

func (e *ScheduleConflictError) Unwrap() error {
	return e.Kind
}

func newScheduleConflictError(kind error, date, startTime string, durationMinutes int64, status, teacherName, studentName string) *ScheduleConflictError {
	endTime := ""
	if startTime != "" {
		endTime = utils.EndTimeFromStartAndDuration(startTime, durationMinutes)
	}
	return &ScheduleConflictError{
		Kind: kind,
		Conflict: ScheduleConflict{
			Date:        date,
			StartTime:   formatTimeHM(startTime),
			EndTime:     formatTimeHM(endTime),
			Status:      conflictStatusLabel(status),
			TeacherName: teacherName,
			StudentName: studentName,
		},
	}
}

func conflictStatusLabel(status string) string {
	if constants.ValidClassStatus(status) {
		return constants.ClassStatus(status).Label()
	}
	if status == string(constants.ScheduledClassStatusScheduled) {
		return "Scheduled"
	}
	if status == "" {
		return "-"
	}
	return status
}

func formatTimeHM(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := utils.ParseTimeHM(value)
	if err != nil {
		return value
	}
	return parsed.Format("15:04")
}

func conflictEndTime(startTime sql.NullString, endTime sql.NullString, durationMinutes int64) string {
	if endTime.Valid && endTime.String != "" {
		return formatTimeHM(endTime.String)
	}
	if startTime.Valid && startTime.String != "" {
		return formatTimeHM(utils.EndTimeFromStartAndDuration(startTime.String, durationMinutes))
	}
	return ""
}

type scheduledClassDB interface {
	GetStudentByID(ctx context.Context, id int64) (queries.GetStudentByIDRow, error)
	CountClassRecordDuplicate(ctx context.Context, arg queries.CountClassRecordDuplicateParams) (int64, error)
	GetClassRecordDuplicate(ctx context.Context, arg queries.GetClassRecordDuplicateParams) (queries.GetClassRecordDuplicateRow, error)
	CountScheduledDuplicate(ctx context.Context, arg queries.CountScheduledDuplicateParams) (int64, error)
	GetScheduledDuplicate(ctx context.Context, arg queries.GetScheduledDuplicateParams) (queries.GetScheduledDuplicateRow, error)
	GetScheduledClassesByTeacherOnDate(ctx context.Context, arg queries.GetScheduledClassesByTeacherOnDateParams) ([]queries.GetScheduledClassesByTeacherOnDateRow, error)
	GetScheduledClassesByStudentOnDate(ctx context.Context, arg queries.GetScheduledClassesByStudentOnDateParams) ([]queries.GetScheduledClassesByStudentOnDateRow, error)
	IsStudentAssignedToTeacher(ctx context.Context, arg queries.IsStudentAssignedToTeacherParams) (int64, error)
}

type ScheduledClassInput struct {
	ScheduleID      int64
	StudentID       int64
	TeacherID       int64
	Date            string
	StartTime       string
	DurationMinutes int64
}

type ScheduledClassRules struct {
	DB scheduledClassDB
}

func (r ScheduledClassRules) Validate(ctx context.Context, actor auth.User, input ScheduledClassInput) error {
	student, err := r.DB.GetStudentByID(ctx, input.StudentID)
	if err != nil {
		return ErrStudentNotFound
	}
	if student.Status != "active" {
		return ErrInactiveStudent
	}

	if actor.Role == auth.RoleTeacher {
		if input.TeacherID != actor.ID {
			return ErrTeacherNotOwner
		}
	}

	assigned, err := r.DB.IsStudentAssignedToTeacher(ctx, queries.IsStudentAssignedToTeacherParams{
		TeacherID: input.TeacherID,
		StudentID: input.StudentID,
	})
	if err != nil {
		return ErrVerifyStudentAssignment
	}
	if assigned == 0 {
		return ErrStudentNotAssigned
	}

	if input.StartTime != "" {
		if err := r.validateNoScheduleOverlap(ctx, input); err != nil {
			return err
		}
	}

	if err := r.validateDuplicateScheduled(ctx, input); err != nil {
		return err
	}

	if err := r.validateDuplicateClassRecord(ctx, input); err != nil {
		return err
	}

	return nil
}

func (r ScheduledClassRules) validateDuplicateScheduled(ctx context.Context, input ScheduledClassInput) error {
	dup, err := r.DB.CountScheduledDuplicate(ctx, queries.CountScheduledDuplicateParams{
		StudentID:       input.StudentID,
		TeacherID:       input.TeacherID,
		ScheduledDate:   input.Date,
		DurationMinutes: input.DurationMinutes,
		Column5:         input.ScheduleID,
		ID:              input.ScheduleID,
	})
	if err != nil {
		return ErrCheckScheduledDuplicate
	}
	if dup == 0 {
		return nil
	}

	row, err := r.DB.GetScheduledDuplicate(ctx, queries.GetScheduledDuplicateParams{
		StudentID:       input.StudentID,
		TeacherID:       input.TeacherID,
		ScheduledDate:   input.Date,
		DurationMinutes: input.DurationMinutes,
		Column5:         input.ScheduleID,
		ID:              input.ScheduleID,
	})
	if err != nil {
		return ErrDuplicateScheduled
	}

	startTime := ""
	if row.StartTime.Valid {
		startTime = row.StartTime.String
	}
	return newScheduleConflictError(
		ErrDuplicateScheduled,
		row.ScheduledDate,
		startTime,
		row.DurationMinutes,
		row.Status,
		row.TeacherName,
		row.StudentName,
	)
}

func (r ScheduledClassRules) validateDuplicateClassRecord(ctx context.Context, input ScheduledClassInput) error {
	dup, err := r.DB.CountClassRecordDuplicate(ctx, queries.CountClassRecordDuplicateParams{
		StudentID:       input.StudentID,
		TeacherID:       input.TeacherID,
		Date:            input.Date,
		DurationMinutes: input.DurationMinutes,
		Column5:         0,
		ID:              0,
	})
	if err != nil {
		return ErrCheckClassRecordDuplicate
	}
	if dup == 0 {
		return nil
	}

	row, err := r.DB.GetClassRecordDuplicate(ctx, queries.GetClassRecordDuplicateParams{
		StudentID:       input.StudentID,
		TeacherID:       input.TeacherID,
		Date:            input.Date,
		DurationMinutes: input.DurationMinutes,
		Column5:         0,
		ID:              0,
	})
	if err != nil {
		return ErrDuplicateClass
	}

	startTime := ""
	if row.StartTime.Valid {
		startTime = row.StartTime.String
	}
	conflict := ScheduleConflict{
		Date:        row.Date,
		StartTime:   formatTimeHM(startTime),
		EndTime:     conflictEndTime(row.StartTime, row.EndTime, row.DurationMinutes),
		Status:      conflictStatusLabel(row.Status),
		TeacherName: row.TeacherName,
		StudentName: row.StudentName,
	}
	return &ScheduleConflictError{
		Kind:     ErrDuplicateClass,
		Conflict: conflict,
	}
}

func (r ScheduledClassRules) validateNoScheduleOverlap(ctx context.Context, input ScheduledClassInput) error {
	excludeID := input.ScheduleID

	teacherClasses, err := r.DB.GetScheduledClassesByTeacherOnDate(ctx, queries.GetScheduledClassesByTeacherOnDateParams{
		TeacherID:     input.TeacherID,
		ScheduledDate: input.Date,
		Column3:       excludeID,
		ID:            excludeID,
	})
	if err != nil {
		return ErrCheckTeacherScheduleConflict
	}
	if row, ok := findTeacherScheduleOverlap(teacherClasses, input.StartTime, input.DurationMinutes); ok {
		startTime := ""
		if row.StartTime.Valid {
			startTime = row.StartTime.String
		}
		return newScheduleConflictError(
			ErrTeacherScheduleConflict,
			row.ScheduledDate,
			startTime,
			row.DurationMinutes,
			row.Status,
			row.TeacherName,
			row.StudentName,
		)
	}

	studentClasses, err := r.DB.GetScheduledClassesByStudentOnDate(ctx, queries.GetScheduledClassesByStudentOnDateParams{
		StudentID:     input.StudentID,
		ScheduledDate: input.Date,
		Column3:       excludeID,
		ID:            excludeID,
	})
	if err != nil {
		return ErrCheckStudentScheduleConflict
	}
	if row, ok := findStudentScheduleOverlap(studentClasses, input.StartTime, input.DurationMinutes); ok {
		startTime := ""
		if row.StartTime.Valid {
			startTime = row.StartTime.String
		}
		return newScheduleConflictError(
			ErrStudentScheduleConflict,
			row.ScheduledDate,
			startTime,
			row.DurationMinutes,
			row.Status,
			row.TeacherName,
			row.StudentName,
		)
	}

	return nil
}

func findTeacherScheduleOverlap(rows []queries.GetScheduledClassesByTeacherOnDateRow, startTime string, durationMinutes int64) (queries.GetScheduledClassesByTeacherOnDateRow, bool) {
	startMins, err := utils.MinutesSinceMidnight(startTime)
	if err != nil {
		return queries.GetScheduledClassesByTeacherOnDateRow{}, false
	}
	for _, row := range rows {
		if !row.StartTime.Valid {
			continue
		}
		existingStart, err := utils.MinutesSinceMidnight(row.StartTime.String)
		if err != nil {
			continue
		}
		if utils.TimeRangesOverlapMinutes(existingStart, row.DurationMinutes, startMins, durationMinutes) {
			return row, true
		}
	}
	return queries.GetScheduledClassesByTeacherOnDateRow{}, false
}

func findStudentScheduleOverlap(rows []queries.GetScheduledClassesByStudentOnDateRow, startTime string, durationMinutes int64) (queries.GetScheduledClassesByStudentOnDateRow, bool) {
	startMins, err := utils.MinutesSinceMidnight(startTime)
	if err != nil {
		return queries.GetScheduledClassesByStudentOnDateRow{}, false
	}
	for _, row := range rows {
		if !row.StartTime.Valid {
			continue
		}
		existingStart, err := utils.MinutesSinceMidnight(row.StartTime.String)
		if err != nil {
			continue
		}
		if utils.TimeRangesOverlapMinutes(existingStart, row.DurationMinutes, startMins, durationMinutes) {
			return row, true
		}
	}
	return queries.GetScheduledClassesByStudentOnDateRow{}, false
}

func (r ScheduledClassRules) ValidateAccess(scheduleTeacherID int64, actor auth.User) error {
	if actor.Role == auth.RoleTeacher && scheduleTeacherID != actor.ID {
		return ErrScheduleNotOwner
	}
	return nil
}
