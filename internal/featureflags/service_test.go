package featureflags

import (
	"testing"
	"zion-english/internal/constants"
)

func TestParseVisibleRolesDefaults(t *testing.T) {
	roles := ParseVisibleRoles("")
	if len(roles) != 4 {
		t.Fatalf("expected 4 default roles, got %d", len(roles))
	}
}

func TestParseVisibleRolesIncludesTester(t *testing.T) {
	roles := ParseVisibleRoles("teacher,tester")
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[0] != constants.TeacherRoleTeacher || roles[1] != constants.TeacherRoleTester {
		t.Fatalf("unexpected roles: %v", roles)
	}
}

func TestParseVisibleRolesFiltersInvalid(t *testing.T) {
	roles := ParseVisibleRoles("teacher,invalid,developer")
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[0] != constants.TeacherRoleTeacher || roles[1] != constants.TeacherRoleDeveloper {
		t.Fatalf("unexpected roles: %v", roles)
	}
}

func TestParseVisibleRolesDedupes(t *testing.T) {
	roles := ParseVisibleRoles("teacher,teacher,developer")
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
}

func TestFormatVisibleRoles(t *testing.T) {
	got := FormatVisibleRoles([]constants.TeacherRole{
		constants.TeacherRoleTeacher,
		constants.TeacherRoleDeveloper,
	})
	if got != "teacher,developer" {
		t.Fatalf("unexpected formatted roles: %q", got)
	}
}

func TestFormatVisibleRolesEmpty(t *testing.T) {
	if FormatVisibleRoles(nil) != "" {
		t.Fatal("expected empty string for nil roles")
	}
}
