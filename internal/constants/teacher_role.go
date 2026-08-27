package constants

type TeacherRole string

const (
	TeacherRoleTeacher   TeacherRole = "teacher"
	TeacherRoleAdmin     TeacherRole = "admin"
	TeacherRoleDeveloper TeacherRole = "developer"
)

var TeacherRoleDisplayPriority = []TeacherRole{
	TeacherRoleAdmin,
	TeacherRoleDeveloper,
	TeacherRoleTeacher,
}

func AllTeacherRoles() []TeacherRole {
	return []TeacherRole{
		TeacherRoleTeacher,
		TeacherRoleAdmin,
		TeacherRoleDeveloper,
	}
}
