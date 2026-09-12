package learningmaterials

import "zion-english/internal/auth"

func CanView(user auth.User, ownerID int64, status, access string) bool {
	if user.Role == auth.RoleSuperuser {
		return true
	}
	if ownerID == user.ID {
		return true
	}
	if status == StatusDeleted {
		return false
	}
	return status == StatusPublished && access == AccessPublic
}

func CanEdit(user auth.User, ownerID int64, status string) bool {
	if user.Role == auth.RoleSuperuser {
		return true
	}
	return ownerID == user.ID && user.ID != 0
}

func CanDelete(user auth.User) bool {
	return user.Role == auth.RoleSuperuser
}
