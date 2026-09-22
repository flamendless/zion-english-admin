package trainingmaterials

import "zion-english/internal/auth"

func CanCreate(role auth.Role) bool {
	return auth.HasAdminAccess(role)
}

func CanEdit(role auth.Role) bool {
	return auth.HasAdminAccess(role)
}

func CanDelete(role auth.Role) bool {
	return auth.HasAdminAccess(role)
}

func CanView(role auth.Role, status string) bool {
	if auth.HasAdminAccess(role) {
		return true
	}
	return status == StatusPublished
}

func CanWatch(role auth.Role, status string) bool {
	if status != StatusPublished {
		return false
	}
	return role == auth.RoleTeacher || auth.HasAdminAccess(role)
}
