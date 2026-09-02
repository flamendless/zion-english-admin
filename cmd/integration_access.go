package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/featureflags"
)

var ErrFeatureFlagRoleRequired = errors.New("at least one role must be selected")

func integrationAccessForViewer(ctx context.Context, key constants.FeatureFlagKey) (visible bool, connectAllowed bool) {
	viewer := auth.GetUser(ctx)
	if auth.HasAdminAccess(viewer.Role) {
		return true, featureflags.IsEnabled(ctx, dbRO, key)
	}
	roles, err := loadTeacherRoles(ctx, viewer.ID)
	if err != nil {
		return false, false
	}
	visible = featureflags.IsVisibleToTeacherRoles(ctx, dbRO, key, roles)
	connectAllowed = featureflags.CanConnect(ctx, dbRO, key, roles)
	return visible, connectAllowed
}

func featureFlagRoleFieldName(prefix string, role constants.TeacherRole) string {
	return fmt.Sprintf("%s_role_%s", prefix, role)
}

func parseFeatureFlagRolesFromForm(r *http.Request, prefix string) ([]constants.TeacherRole, error) {
	roles := make([]constants.TeacherRole, 0, len(constants.AllTeacherRoles()))
	for _, role := range constants.AllTeacherRoles() {
		if r.FormValue(featureFlagRoleFieldName(prefix, role)) == "on" {
			roles = append(roles, role)
		}
	}
	if len(roles) == 0 {
		return nil, ErrFeatureFlagRoleRequired
	}
	return roles, nil
}

func formatVisibleRolesAudit(roles []constants.TeacherRole) string {
	parts := make([]string, len(roles))
	for i, role := range roles {
		parts[i] = string(role)
	}
	return strings.Join(parts, ", ")
}
