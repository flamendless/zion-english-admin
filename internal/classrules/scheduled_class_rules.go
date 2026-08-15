package classrules

import (
	"context"
	"errors"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
)

var (
	ErrDuplicateScheduled = errors.New("a scheduled class with the same student, teacher, date, and duration already exists")
	ErrScheduleNotOwner   = errors.New("you can only manage your own scheduled classes")
)

type scheduledClassDB interface {
	GetStudentByID(ctx context.Context, id int64) (queries.TblStudent, error)
	CountClassRecordDuplicate(ctx context.Context, arg queries.CountClassRecordDuplicateParams) (int64, error)
	CountScheduledDuplicate(ctx context.Context, arg queries.CountScheduledDuplicateParams) (int64, error)
	IsStudentAssignedToTeacher(ctx context.Context, arg queries.IsStudentAssignedToTeacherParams) (int64, error)
}

type ScheduledClassInput struct {
	ScheduleID      int64
	StudentID       int64
	TeacherID       int64
	Date            string
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

	return nil
}

func (r ScheduledClassRules) ValidateAccess(scheduleTeacherID int64, actor auth.User) error {
	if actor.Role == auth.RoleTeacher && scheduleTeacherID != actor.ID {
		return ErrScheduleNotOwner
	}
	return nil
}
