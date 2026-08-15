package classrules_test

import (
	"context"
	"errors"
	"testing"

	"zion-english/internal/auth"
	"zion-english/internal/classrules"
	"zion-english/internal/database/queries"
)

type mockScheduleDB struct {
	mockDB
	scheduledDup int64
}

func (m *mockScheduleDB) CountScheduledDuplicate(ctx context.Context, arg queries.CountScheduledDuplicateParams) (int64, error) {
	return m.scheduledDup, nil
}

func TestValidateDuplicateScheduled(t *testing.T) {
	rules := classrules.ScheduledClassRules{DB: &mockScheduleDB{
		mockDB: mockDB{
			student:  queries.TblStudent{ID: 1, Status: "active"},
			assigned: 1,
		},
		scheduledDup: 1,
	}}
	err := rules.Validate(context.Background(), auth.User{Role: auth.RoleSuperuser}, classrules.ScheduledClassInput{
		StudentID: 1, TeacherID: 2, Date: "2026-01-01", DurationMinutes: 60,
	})
	if !errors.Is(err, classrules.ErrDuplicateScheduled) {
		t.Fatalf("expected ErrDuplicateScheduled, got %v", err)
	}
}

func TestValidateScheduleAccess(t *testing.T) {
	rules := classrules.ScheduledClassRules{DB: &mockScheduleDB{}}
	err := rules.ValidateAccess(2, auth.User{ID: 1, Role: auth.RoleTeacher})
	if !errors.Is(err, classrules.ErrScheduleNotOwner) {
		t.Fatalf("expected ErrScheduleNotOwner, got %v", err)
	}
}
