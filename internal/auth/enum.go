package auth

import "context"

type Role string

const (
	RoleTeacher   Role = "teacher"
	RoleAdmin     Role = "admin"
	RoleTester    Role = "tester"
	RoleSuperuser Role = "superuser"
)

func AdminAccessRoles() []Role {
	return []Role{RoleSuperuser, RoleAdmin}
}

func HasAdminAccess(role Role) bool {
	return role == RoleSuperuser || role == RoleAdmin
}

func IsTeacherScoped(role Role) bool {
	return role == RoleTeacher || role == RoleTester
}

type User struct {
	ID    int64
	Name  string
	Email string
	Role  Role
}

type contextKey string

const userKey contextKey = "user"

func GetUser(ctx context.Context) User {
	user, _ := ctx.Value(userKey).(User)
	return user
}

func GetRole(ctx context.Context) Role {
	user := GetUser(ctx)
	return user.Role
}
