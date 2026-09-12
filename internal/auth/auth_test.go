package auth

import (
	"testing"
	"zion-english/internal/constants"
)

func TestResolveLoginRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []constants.TeacherRole
		want  Role
	}{
		{
			name:  "admin wins",
			roles: []constants.TeacherRole{constants.TeacherRoleTeacher, constants.TeacherRoleAdmin},
			want:  RoleAdmin,
		},
		{
			name:  "tester without admin",
			roles: []constants.TeacherRole{constants.TeacherRoleTester},
			want:  RoleTester,
		},
		{
			name:  "teacher default",
			roles: []constants.TeacherRole{constants.TeacherRoleTeacher, constants.TeacherRoleDeveloper},
			want:  RoleTeacher,
		},
		{
			name:  "tester over teacher",
			roles: []constants.TeacherRole{constants.TeacherRoleTeacher, constants.TeacherRoleTester},
			want:  RoleTester,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveLoginRole(tt.roles); got != tt.want {
				t.Fatalf("resolveLoginRole(%v) = %q, want %q", tt.roles, got, tt.want)
			}
		})
	}
}

func TestIsTeacherScoped(t *testing.T) {
	if !IsTeacherScoped(RoleTeacher) || !IsTeacherScoped(RoleTester) {
		t.Fatal("teacher and tester should be teacher-scoped")
	}
	if IsTeacherScoped(RoleAdmin) || IsTeacherScoped(RoleSuperuser) {
		t.Fatal("admin and superuser should not be teacher-scoped")
	}
}
