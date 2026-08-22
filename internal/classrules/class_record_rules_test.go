package classrules_test

import (
	"context"
	"errors"
	"testing"

	"zion-english/internal/auth"
	"zion-english/internal/classrules"
	"zion-english/internal/database/queries"
)

type mockDB struct {
	student      queries.GetStudentByIDRow
	studentErr   error
	duplicate    int64
	duplicateErr error
	assigned     int64
	assignedErr  error
}

func (m *mockDB) GetStudentByID(ctx context.Context, id int64) (queries.GetStudentByIDRow, error) {
	return m.student, m.studentErr
}

func (m *mockDB) CountClassRecordDuplicate(ctx context.Context, arg queries.CountClassRecordDuplicateParams) (int64, error) {
	return m.duplicate, m.duplicateErr
}

func (m *mockDB) IsStudentAssignedToTeacher(ctx context.Context, arg queries.IsStudentAssignedToTeacherParams) (int64, error) {
	return m.assigned, m.assignedErr
}

func TestValidateInactiveStudent(t *testing.T) {
	rules := classrules.ClassRecordRules{DB: &mockDB{
		student: queries.GetStudentByIDRow{ID: 1, Status: "inactive"},
	}}
	err := rules.Validate(context.Background(), auth.User{Role: auth.RoleSuperuser}, classrules.ClassRecordInput{
		StudentID: 1, TeacherID: 2, Date: "2026-01-01", DurationMinutes: 60,
	})
	if !errors.Is(err, classrules.ErrInactiveStudent) {
		t.Fatalf("expected ErrInactiveStudent, got %v", err)
	}
}

func TestValidateDuplicateClass(t *testing.T) {
	rules := classrules.ClassRecordRules{DB: &mockDB{
		student:   queries.GetStudentByIDRow{ID: 1, Status: "active"},
		duplicate: 1,
	}}
	err := rules.Validate(context.Background(), auth.User{Role: auth.RoleSuperuser}, classrules.ClassRecordInput{
		StudentID: 1, TeacherID: 2, Date: "2026-01-01", DurationMinutes: 60,
	})
	if !errors.Is(err, classrules.ErrDuplicateClass) {
		t.Fatalf("expected ErrDuplicateClass, got %v", err)
	}
}

func TestValidateTeacherAssignment(t *testing.T) {
	rules := classrules.ClassRecordRules{DB: &mockDB{
		student: queries.GetStudentByIDRow{ID: 1, Status: "active"},
	}}
	err := rules.Validate(context.Background(), auth.User{ID: 5, Role: auth.RoleTeacher}, classrules.ClassRecordInput{
		StudentID: 1, TeacherID: 5, Date: "2026-01-01", DurationMinutes: 60,
	})
	if !errors.Is(err, classrules.ErrStudentNotAssigned) {
		t.Fatalf("expected ErrStudentNotAssigned, got %v", err)
	}
}
