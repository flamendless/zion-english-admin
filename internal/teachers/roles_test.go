package teachers

import (
	"testing"
	"zion-english/internal/constants"
)

func TestCanManageTeacherRoles(t *testing.T) {
	if !CanManageTeacherRoles("superuser", true) {
		t.Fatal("superuser should manage roles on admin-tagged teachers")
	}
	if CanManageTeacherRoles("admin", true) {
		t.Fatal("admin should not manage roles on admin-tagged teachers")
	}
	if !CanManageTeacherRoles("admin", false) {
		t.Fatal("admin should manage roles on non-admin teachers")
	}
	if CanManageTeacherRoles("teacher", false) {
		t.Fatal("teacher should not manage roles")
	}
}

func TestParseTeacherRolesRequiresTeacher(t *testing.T) {
	_, err := ParseTeacherRoles([]string{"developer"})
	if err != ErrTeacherRoleRequired {
		t.Fatalf("expected ErrTeacherRoleRequired, got %v", err)
	}

	roles, err := ParseTeacherRoles([]string{"teacher", "developer", "developer"})
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
}

func TestValidateRoleAssignment(t *testing.T) {
	submitted := []constants.TeacherRole{constants.TeacherRoleTeacher, constants.TeacherRoleDeveloper}
	if err := ValidateRoleAssignment("admin", true, submitted); err != ErrRolesForbidden {
		t.Fatalf("expected ErrRolesForbidden, got %v", err)
	}
	if err := ValidateRoleAssignment("admin", false, submitted); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestPrimaryTeacherRole(t *testing.T) {
	role, ok := PrimaryTeacherRole([]constants.TeacherRole{constants.TeacherRoleTeacher})
	if !ok || role != constants.TeacherRoleTeacher {
		t.Fatalf("expected teacher, got %q ok=%v", role, ok)
	}

	role, ok = PrimaryTeacherRole([]constants.TeacherRole{
		constants.TeacherRoleTeacher,
		constants.TeacherRoleDeveloper,
	})
	if !ok || role != constants.TeacherRoleDeveloper {
		t.Fatalf("expected developer, got %q ok=%v", role, ok)
	}

	role, ok = PrimaryTeacherRole([]constants.TeacherRole{
		constants.TeacherRoleTeacher,
		constants.TeacherRoleDeveloper,
		constants.TeacherRoleAdmin,
	})
	if !ok || role != constants.TeacherRoleAdmin {
		t.Fatalf("expected admin, got %q ok=%v", role, ok)
	}
}
