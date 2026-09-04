package classrules_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"zion-english/internal/auth"
	"zion-english/internal/classrules"
	"zion-english/internal/database/queries"
)

type mockScheduleDB struct {
	mockDB
	scheduledDup       int64
	scheduledDuplicate queries.GetScheduledDuplicateRow
	teacherSchedules   []queries.GetScheduledClassesByTeacherOnDateRow
	studentSchedules   []queries.GetScheduledClassesByStudentOnDateRow
}

func (m *mockScheduleDB) CountScheduledDuplicate(ctx context.Context, arg queries.CountScheduledDuplicateParams) (int64, error) {
	return m.scheduledDup, nil
}

func (m *mockScheduleDB) GetScheduledDuplicate(ctx context.Context, arg queries.GetScheduledDuplicateParams) (queries.GetScheduledDuplicateRow, error) {
	return m.scheduledDuplicate, nil
}

func (m *mockScheduleDB) GetClassRecordDuplicate(ctx context.Context, arg queries.GetClassRecordDuplicateParams) (queries.GetClassRecordDuplicateRow, error) {
	return queries.GetClassRecordDuplicateRow{}, sql.ErrNoRows
}

func (m *mockScheduleDB) GetScheduledClassesByTeacherOnDate(ctx context.Context, arg queries.GetScheduledClassesByTeacherOnDateParams) ([]queries.GetScheduledClassesByTeacherOnDateRow, error) {
	return m.teacherSchedules, nil
}

func (m *mockScheduleDB) GetScheduledClassesByStudentOnDate(ctx context.Context, arg queries.GetScheduledClassesByStudentOnDateParams) ([]queries.GetScheduledClassesByStudentOnDateRow, error) {
	return m.studentSchedules, nil
}

func activeScheduleStudent() queries.GetStudentByIDRow {
	return queries.GetStudentByIDRow{ID: 1, Status: "active"}
}

func scheduleStartTime(value string) sql.NullString {
	return sql.NullString{String: value, Valid: true}
}

func TestValidateDuplicateScheduled(t *testing.T) {
	rules := classrules.ScheduledClassRules{DB: &mockScheduleDB{
		mockDB: mockDB{
			student:  activeScheduleStudent(),
			assigned: 1,
		},
		scheduledDup: 1,
		scheduledDuplicate: queries.GetScheduledDuplicateRow{
			ScheduledDate:   "2026-01-01",
			StartTime:       scheduleStartTime("10:00"),
			DurationMinutes: 60,
			Status:          "scheduled",
			TeacherName:     "Jane Teacher",
			StudentName:     "John Student",
		},
	}}
	err := rules.Validate(context.Background(), auth.User{Role: auth.RoleSuperuser}, classrules.ScheduledClassInput{
		StudentID: 1, TeacherID: 2, Date: "2026-01-01", StartTime: "10:00", DurationMinutes: 60,
	})
	if !errors.Is(err, classrules.ErrDuplicateScheduled) {
		t.Fatalf("expected ErrDuplicateScheduled, got %v", err)
	}
	if err.Error() != "a scheduled class with the same student, teacher, date, and duration already exists. Conflicting class: 2026-01-01, 10:00 - 11:00, status: Scheduled, teacher: Jane Teacher, student: John Student" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateTeacherScheduleConflict(t *testing.T) {
	rules := classrules.ScheduledClassRules{DB: &mockScheduleDB{
		mockDB: mockDB{
			student:  activeScheduleStudent(),
			assigned: 1,
		},
		teacherSchedules: []queries.GetScheduledClassesByTeacherOnDateRow{
			{
				ID:              9,
				ScheduledDate:   "2026-01-01",
				StartTime:       scheduleStartTime("10:00"),
				DurationMinutes: 60,
				Status:          "scheduled",
				TeacherName:     "Jane Teacher",
				StudentName:     "John Student",
			},
		},
	}}
	err := rules.Validate(context.Background(), auth.User{Role: auth.RoleSuperuser}, classrules.ScheduledClassInput{
		StudentID: 1, TeacherID: 2, Date: "2026-01-01", StartTime: "10:30", DurationMinutes: 60,
	})
	if !errors.Is(err, classrules.ErrTeacherScheduleConflict) {
		t.Fatalf("expected ErrTeacherScheduleConflict, got %v", err)
	}
	if err.Error() != "teacher already has a class scheduled at this time. Conflicting class: 2026-01-01, 10:00 - 11:00, status: Scheduled, teacher: Jane Teacher, student: John Student" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateStudentScheduleConflict(t *testing.T) {
	rules := classrules.ScheduledClassRules{DB: &mockScheduleDB{
		mockDB: mockDB{
			student:  activeScheduleStudent(),
			assigned: 1,
		},
		studentSchedules: []queries.GetScheduledClassesByStudentOnDateRow{
			{
				ID:              9,
				ScheduledDate:   "2026-01-01",
				StartTime:       scheduleStartTime("14:00"),
				DurationMinutes: 30,
				Status:          "scheduled",
				TeacherName:     "Jane Teacher",
				StudentName:     "John Student",
			},
		},
	}}
	err := rules.Validate(context.Background(), auth.User{Role: auth.RoleSuperuser}, classrules.ScheduledClassInput{
		StudentID: 1, TeacherID: 2, Date: "2026-01-01", StartTime: "14:15", DurationMinutes: 30,
	})
	if !errors.Is(err, classrules.ErrStudentScheduleConflict) {
		t.Fatalf("expected ErrStudentScheduleConflict, got %v", err)
	}
	if err.Error() != "student already has a class scheduled at this time. Conflicting class: 2026-01-01, 14:00 - 14:30, status: Scheduled, teacher: Jane Teacher, student: John Student" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestValidateAdjacentSchedulesAllowed(t *testing.T) {
	rules := classrules.ScheduledClassRules{DB: &mockScheduleDB{
		mockDB: mockDB{
			student:  activeScheduleStudent(),
			assigned: 1,
		},
		teacherSchedules: []queries.GetScheduledClassesByTeacherOnDateRow{
			{ID: 9, ScheduledDate: "2026-01-01", StartTime: scheduleStartTime("10:00"), DurationMinutes: 60},
		},
		studentSchedules: []queries.GetScheduledClassesByStudentOnDateRow{
			{ID: 10, ScheduledDate: "2026-01-01", StartTime: scheduleStartTime("12:00"), DurationMinutes: 60},
		},
	}}
	err := rules.Validate(context.Background(), auth.User{Role: auth.RoleSuperuser}, classrules.ScheduledClassInput{
		StudentID: 1, TeacherID: 2, Date: "2026-01-01", StartTime: "11:00", DurationMinutes: 60,
	})
	if err != nil {
		t.Fatalf("expected no conflict for back-to-back class, got %v", err)
	}
}

func TestValidateScheduleAccess(t *testing.T) {
	rules := classrules.ScheduledClassRules{DB: &mockScheduleDB{}}
	err := rules.ValidateAccess(2, auth.User{ID: 1, Role: auth.RoleTeacher})
	if !errors.Is(err, classrules.ErrScheduleNotOwner) {
		t.Fatalf("expected ErrScheduleNotOwner, got %v", err)
	}
}
