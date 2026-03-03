package auth

import "context"

type Role string

const (
	RoleTeacher   Role = "teacher"
	RoleSuperuser Role = "superuser"
)

type contextKey string

const roleKey contextKey = "role"

func GetRole(ctx context.Context) Role {
	role, _ := ctx.Value(roleKey).(Role)
	return role
}
