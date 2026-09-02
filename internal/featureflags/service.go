package featureflags

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
)

func DefaultVisibleRoles() []constants.TeacherRole {
	return slices.Clone(constants.AllTeacherRoles())
}

func FormatVisibleRoles(roles []constants.TeacherRole) string {
	if len(roles) == 0 {
		return ""
	}
	parts := make([]string, len(roles))
	for i, role := range roles {
		parts[i] = string(role)
	}
	return strings.Join(parts, ",")
}

func ParseVisibleRoles(raw string) []constants.TeacherRole {
	if strings.TrimSpace(raw) == "" {
		return DefaultVisibleRoles()
	}
	valid := make(map[constants.TeacherRole]struct{}, len(constants.AllTeacherRoles()))
	for _, role := range constants.AllTeacherRoles() {
		valid[role] = struct{}{}
	}
	seen := make(map[constants.TeacherRole]struct{})
	roles := make([]constants.TeacherRole, 0)
	for _, part := range strings.Split(raw, ",") {
		role := constants.TeacherRole(strings.TrimSpace(part))
		if role == "" {
			continue
		}
		if _, ok := valid[role]; !ok {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return DefaultVisibleRoles()
	}
	return roles
}

func GetFlag(ctx context.Context, db database.Service, key constants.FeatureFlagKey) (enabled bool, visibleRoles []constants.TeacherRole, err error) {
	row, err := db.GetQueries().GetFeatureFlag(ctx, string(key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, DefaultVisibleRoles(), nil
		}
		return true, DefaultVisibleRoles(), err
	}
	return row.Enabled == 1, ParseVisibleRoles(row.VisibleRoles), nil
}

func IsEnabled(ctx context.Context, db database.Service, key constants.FeatureFlagKey) bool {
	enabled, _, err := GetFlag(ctx, db, key)
	if err != nil {
		return true
	}
	return enabled
}

func VisibleRoles(ctx context.Context, db database.Service, key constants.FeatureFlagKey) []constants.TeacherRole {
	_, roles, err := GetFlag(ctx, db, key)
	if err != nil {
		return DefaultVisibleRoles()
	}
	return roles
}

func IsVisibleToTeacherRoles(ctx context.Context, db database.Service, key constants.FeatureFlagKey, roles []constants.TeacherRole) bool {
	allowed := VisibleRoles(ctx, db, key)
	for _, role := range roles {
		if slices.Contains(allowed, role) {
			return true
		}
	}
	return false
}

func CanConnect(ctx context.Context, db database.Service, key constants.FeatureFlagKey, roles []constants.TeacherRole) bool {
	return IsEnabled(ctx, db, key) && IsVisibleToTeacherRoles(ctx, db, key, roles)
}

func SetFlag(ctx context.Context, db database.Service, key constants.FeatureFlagKey, enabled bool, visibleRoles []constants.TeacherRole) error {
	var enabledInt int64
	if enabled {
		enabledInt = 1
	}
	return db.GetQueries().UpsertFeatureFlag(ctx, queries.UpsertFeatureFlagParams{
		Key:          string(key),
		Enabled:      enabledInt,
		VisibleRoles: FormatVisibleRoles(visibleRoles),
	})
}

func SetEnabled(ctx context.Context, db database.Service, key constants.FeatureFlagKey, enabled bool) error {
	visibleRoles := VisibleRoles(ctx, db, key)
	return SetFlag(ctx, db, key, enabled, visibleRoles)
}
