package frontend

import "zion-english/internal/auth"

func IsNavAccessible(role auth.Role, path string) bool {
	if role == "" {
		return false
	}

	if role == auth.RoleSuperuser {
		if path == "/my-students" {
			return false
		}
		return true
	}

	switch path {
	case "/classes", "/classes/record", "/profile", "/logs", "/my-students":
		return true
	default:
		return false
	}
}