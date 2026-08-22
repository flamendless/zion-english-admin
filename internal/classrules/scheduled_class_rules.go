package classrules

import (
	"context"
	"errors"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

var (
	ErrDuplicateScheduled      = errors.New("a scheduled class with the same student, teacher, date, and duration already exists")
	ErrScheduleNotOwner        = errors.New("you can only manage your own scheduled classes")
	ErrTeacherScheduleConflict = errors.New("teacher already has a class scheduled at this time")
	ErrStudentScheduleConflict = errors.New("student already has a class scheduled at this time")
)

type scheduledClassTimeSlot struct {
	StartTime       string
	DurationMinutes int64
}

type scheduledClassDB interface {
	GetStudentByID(ctx context.Context, id int64) (queries.GetStudentByIDRow, error)
	CountClassRecordDuplicate(ctx context.Context, arg queries.CountClassRecordDuplicateParams) (int64, error)
	CountScheduledDuplicate(ctx context.Context, arg queries.CountScheduledDuplicateParams) (int64, error)
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

	dup, err := r.DB.CountScheduledDuplicate(ctx, queries.CountScheduledDuplicateParams{
		StudentID:       input.StudentID,
		TeacherID:       input.TeacherID,
		ScheduledDate:   input.Date,
		DurationMinutes: input.DurationMinutes,
		Column5:         input.ScheduleID,
		ID:              input.ScheduleID,
	})
	if err != nil {
		return errors.New("failed to check duplicate scheduled class")
	}
	if dup > 0 {
		return ErrDuplicateScheduled
	}

	recordDup, err := r.DB.CountClassRecordDuplicate(ctx, queries.CountClassRecordDuplicateParams{
		StudentID:       input.StudentID,
		TeacherID:       input.TeacherID,
		Date:            input.Date,
		DurationMinutes: input.DurationMinutes,
		Column5:         0,
		ID:              0,
	})
	if err != nil {
		return errors.New("failed to check duplicate class record")
	}
	if recordDup > 0 {
		return ErrDuplicateClass
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
		return errors.New("failed to verify student assignment")
	}
	if assigned == 0 {
		return ErrStudentNotAssigned
	}

	if input.StartTime != "" {
		if err := r.validateNoScheduleOverlap(ctx, input); err != nil {
			return err
		}
	}

	return nil
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
		return errors.New("failed to check teacher schedule conflicts")
	}
	if scheduleOverlaps(teacherClassesToSlots(teacherClasses), input.StartTime, input.DurationMinutes) {
		return ErrTeacherScheduleConflict
	}

	studentClasses, err := r.DB.GetScheduledClassesByStudentOnDate(ctx, queries.GetScheduledClassesByStudentOnDateParams{
		StudentID:     input.StudentID,
		ScheduledDate: input.Date,
		Column3:       excludeID,
		ID:            excludeID,
	})
	if err != nil {
		return errors.New("failed to check student schedule conflicts")
	}
	if scheduleOverlaps(studentClassesToSlots(studentClasses), input.StartTime, input.DurationMinutes) {
		return ErrStudentScheduleConflict
	}

	return nil
}

func teacherClassesToSlots(rows []queries.GetScheduledClassesByTeacherOnDateRow) []scheduledClassTimeSlot {
	slots := make([]scheduledClassTimeSlot, 0, len(rows))
	for _, row := range rows {
		if !row.StartTime.Valid {
			continue
		}
		slots = append(slots, scheduledClassTimeSlot{
			StartTime:       row.StartTime.String,
			DurationMinutes: row.DurationMinutes,
		})
	}
	return slots
}

func studentClassesToSlots(rows []queries.GetScheduledClassesByStudentOnDateRow) []scheduledClassTimeSlot {
	slots := make([]scheduledClassTimeSlot, 0, len(rows))
	for _, row := range rows {
		if !row.StartTime.Valid {
			continue
		}
		slots = append(slots, scheduledClassTimeSlot{
			StartTime:       row.StartTime.String,
			DurationMinutes: row.DurationMinutes,
		})
	}
	return slots
}

func scheduleOverlaps(existing []scheduledClassTimeSlot, startTime string, durationMinutes int64) bool {
	startMins, err := utils.MinutesSinceMidnight(startTime)
	if err != nil {
		return false
	}
	for _, slot := range existing {
		existingStart, err := utils.MinutesSinceMidnight(slot.StartTime)
		if err != nil {
			continue
		}
		if utils.TimeRangesOverlapMinutes(existingStart, slot.DurationMinutes, startMins, durationMinutes) {
			return true
		}
	}
	return false
}

func (r ScheduledClassRules) ValidateAccess(scheduleTeacherID int64, actor auth.User) error {
	if actor.Role == auth.RoleTeacher && scheduleTeacherID != actor.ID {
		return ErrScheduleNotOwner
	}
	return nil
}
