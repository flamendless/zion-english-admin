package frontend

import "zion-english/internal/auth"

func IsNavAccessible(role auth.Role, path string) bool {
	if role == "" {
		return false
	}

	if role == auth.RoleSuperuser {
		return true
	}

	switch path {
	case "/classes", "/classes/record":
		return true
	default:
		return false
	}
}
