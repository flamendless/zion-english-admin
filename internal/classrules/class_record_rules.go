package classrules

import (
	"context"
	"errors"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
)

var (
	ErrInactiveStudent    = errors.New("cannot record class for an inactive student")
	ErrDuplicateClass     = errors.New("a class with the same student, teacher, date, and duration already exists")
	ErrStudentNotAssigned = errors.New("student is not assigned to this teacher")
	ErrTeacherNotOwner    = errors.New("you can only edit your own class records")
)

type classRecordDB interface {
	GetStudentByID(ctx context.Context, id int64) (queries.TblStudent, error)
	CountClassRecordDuplicate(ctx context.Context, arg queries.CountClassRecordDuplicateParams) (int64, error)
	IsStudentAssignedToTeacher(ctx context.Context, arg queries.IsStudentAssignedToTeacherParams) (int64, error)
}

type ClassRecordInput struct {
	RecordID        int64
	StudentID       int64
	TeacherID       int64
	Date            string
	DurationMinutes int64
}

type ClassRecordRules struct {
	DB classRecordDB
}

func (r ClassRecordRules) Validate(ctx context.Context, actor auth.User, input ClassRecordInput) error {
	student, err := r.DB.GetStudentByID(ctx, input.StudentID)
	if err != nil {
		return errors.New("student not found")
	}
	if student.Status != "active" {
		return ErrInactiveStudent
	}

	excludeID := input.RecordID
	dup, err := r.DB.CountClassRecordDuplicate(ctx, queries.CountClassRecordDuplicateParams{
		StudentID:       input.StudentID,
		TeacherID:       input.TeacherID,
		Date:            input.Date,
		DurationMinutes: input.DurationMinutes,
		Column5:         excludeID,
		ID:              excludeID,
	})
	if err != nil {
		return errors.New("failed to check duplicate class")
	}
	if dup > 0 {
		return ErrDuplicateClass
	}

	if actor.Role == auth.RoleTeacher {
		if input.TeacherID != actor.ID {
			return ErrTeacherNotOwner
		}
		assigned, err := r.DB.IsStudentAssignedToTeacher(ctx, queries.IsStudentAssignedToTeacherParams{
			TeacherID: actor.ID,
			StudentID: input.StudentID,
		})
		if err != nil {
			return errors.New("failed to verify student assignment")
		}
		if assigned == 0 {
			return ErrStudentNotAssigned
		}
	}

	return nil
}

func (r ClassRecordRules) ValidateEditAccess(recordTeacherID int64, actor auth.User) error {
	if actor.Role == auth.RoleTeacher && recordTeacherID != actor.ID {
		return ErrTeacherNotOwner
	}
	return nil
}
