package ads

import (
	"testing"

	"zion-english/internal/auth"
	"zion-english/internal/constants"
)

func TestSuppressAdsForViewer(t *testing.T) {
	teacher := auth.User{ID: 1, Role: auth.RoleTeacher}
	tests := []struct {
		name  string
		user  auth.User
		roles []constants.TeacherRole
		want  bool
	}{
		{
			name: "tester session role",
			user: auth.User{ID: 2, Role: auth.RoleTester},
			want: true,
		},
		{
			name:  "developer teacher role",
			user:  teacher,
			roles: []constants.TeacherRole{constants.TeacherRoleTeacher, constants.TeacherRoleDeveloper},
			want:  false,
		},
		{
			name:  "tester teacher role on teacher session",
			user:  teacher,
			roles: []constants.TeacherRole{constants.TeacherRoleTester},
			want:  true,
		},
		{
			name:  "regular teacher",
			user:  teacher,
			roles: []constants.TeacherRole{constants.TeacherRoleTeacher},
			want:  false,
		},
		{
			name: "anonymous",
			user: auth.User{},
			want: false,
		},
		{
			name: "superuser",
			user: auth.User{Role: auth.RoleSuperuser},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := suppressAdsForViewer(tc.user, tc.roles); got != tc.want {
				t.Fatalf("suppressAdsForViewer() = %v, want %v", got, tc.want)
			}
		})
	}
}
