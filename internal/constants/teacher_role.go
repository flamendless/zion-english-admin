package constants

type TeacherRole string

const (
	TeacherRoleTeacher   TeacherRole = "teacher"
	TeacherRoleAdmin     TeacherRole = "admin"
	TeacherRoleDeveloper TeacherRole = "developer"
	TeacherRoleTester    TeacherRole = "tester"
)

var TeacherRoleDisplayPriority = []TeacherRole{
	TeacherRoleAdmin,
	TeacherRoleDeveloper,
	TeacherRoleTester,
	TeacherRoleTeacher,
}

func AllTeacherRoles() []TeacherRole {
	return []TeacherRole{
		TeacherRoleTeacher,
		TeacherRoleAdmin,
		TeacherRoleDeveloper,
		TeacherRoleTester,
	}
}
