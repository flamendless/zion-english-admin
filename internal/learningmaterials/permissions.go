package learningmaterials

import (
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
)

func CanView(user auth.User, material queries.TblLearningMaterial) bool {
	if user.Role == auth.RoleSuperuser {
		return true
	}
	if material.Status == StatusDeleted {
		return false
	}
	if material.OwnerID == user.ID {
		return true
	}
	return material.Status == StatusPublished && material.Access == AccessPublic
}

func CanEdit(user auth.User, material queries.TblLearningMaterial) bool {
	if material.Status == StatusDeleted {
		return false
	}
	if user.Role == auth.RoleSuperuser {
		return true
	}
	return material.OwnerID == user.ID && user.ID != 0
}

func CanDelete(user auth.User) bool {
	return user.Role == auth.RoleSuperuser
}
