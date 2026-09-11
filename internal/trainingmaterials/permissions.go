package trainingmaterials

import (
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
)

func CanManage(user auth.User) bool {
	return auth.HasAdminAccess(user.Role)
}

func CanView(user auth.User, material queries.TblTrainingMaterial) bool {
	if auth.HasAdminAccess(user.Role) {
		return true
	}
	return material.Status == string(StatusPublished)
}
