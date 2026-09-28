package ads

import (
	"context"
	"net/http"

	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/teachers"
)

func suppressAdsForViewer(user auth.User, teacherRoles []constants.TeacherRole) bool {
	if user.Role == auth.RoleTester {
		return true
	}
	if user.ID == 0 {
		return false
	}
	return teachers.HasRole(teacherRoles, constants.TeacherRoleTester)
}

func suppressAdsForRequest(ctx context.Context, r *http.Request, db *queries.Queries) bool {
	user, ok := auth.UserFromRequest(r, conf.Conf())
	if !ok {
		return false
	}
	if user.Role == auth.RoleTester {
		return true
	}
	if user.ID == 0 {
		return false
	}
	switch user.Role {
	case auth.RoleTeacher, auth.RoleAdmin:
	default:
		return false
	}
	roleRows, err := db.GetTeacherRolesByTeacherID(ctx, user.ID)
	if err != nil {
		return false
	}
	return suppressAdsForViewer(user, teachers.StringsToRoles(roleRows))
}
