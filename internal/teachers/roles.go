package teachers

import (
	"errors"
	"slices"
	"strings"
	"zion-english/internal/constants"
)

var (
	ErrRolesForbidden      = errors.New("you are not allowed to manage roles for this teacher")
	ErrInvalidRole         = errors.New("invalid role")
	ErrTeacherRoleRequired = errors.New("teacher role is required")
)

func CanManageTeacherRoles(actorRole string, targetHasAdmin bool) bool {
	switch actorRole {
	case "superuser":
		return true
	case "admin":
		return !targetHasAdmin
	default:
		return false
	}
}

func HasRole(roles []constants.TeacherRole, role constants.TeacherRole) bool {
	return slices.Contains(roles, role)
}

func PrimaryTeacherRole(roles []constants.TeacherRole) (constants.TeacherRole, bool) {
	for _, role := range constants.TeacherRoleDisplayPriority {
		if HasRole(roles, role) {
			return role, true
		}
	}
	return "", false
}

func ParseTeacherRoles(values []string) ([]constants.TeacherRole, error) {
	seen := make(map[constants.TeacherRole]struct{}, len(values))
	roles := make([]constants.TeacherRole, 0, len(values))
	for _, raw := range values {
		role := constants.TeacherRole(strings.TrimSpace(raw))
		if role == "" {
			continue
		}
		if !isValidTeacherRole(role) {
			return nil, ErrInvalidRole
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return nil, ErrTeacherRoleRequired
	}
	if !HasRole(roles, constants.TeacherRoleTeacher) && !HasRole(roles, constants.TeacherRoleTester) {
		return nil, ErrTeacherRoleRequired
	}
	return roles, nil
}

func ValidateRoleAssignment(actorRole string, targetHasAdmin bool, submitted []constants.TeacherRole) error {
	if !CanManageTeacherRoles(actorRole, targetHasAdmin) {
		return ErrRolesForbidden
	}
	if len(submitted) == 0 {
		return ErrTeacherRoleRequired
	}
	for _, role := range submitted {
		if !isValidTeacherRole(role) {
			return ErrInvalidRole
		}
	}
	if !HasRole(submitted, constants.TeacherRoleTeacher) && !HasRole(submitted, constants.TeacherRoleTester) {
		return ErrTeacherRoleRequired
	}
	return nil
}

func RolesToStrings(roles []constants.TeacherRole) []string {
	out := make([]string, len(roles))
	for i, role := range roles {
		out[i] = string(role)
	}
	return out
}

func StringsToRoles(values []string) []constants.TeacherRole {
	roles := make([]constants.TeacherRole, len(values))
	for i, value := range values {
		roles[i] = constants.TeacherRole(value)
	}
	return roles
}

func FormatRoleDiff(before, after []constants.TeacherRole) string {
	beforeSet := roleSet(before)
	afterSet := roleSet(after)
	added := make([]string, 0)
	removed := make([]string, 0)
	for role := range afterSet {
		if _, ok := beforeSet[role]; !ok {
			added = append(added, string(role))
		}
	}
	for role := range beforeSet {
		if _, ok := afterSet[role]; !ok {
			removed = append(removed, string(role))
		}
	}
	slices.Sort(added)
	slices.Sort(removed)
	parts := make([]string, 0, 2)
	if len(added) > 0 {
		parts = append(parts, "added "+strings.Join(added, ", "))
	}
	if len(removed) > 0 {
		parts = append(parts, "removed "+strings.Join(removed, ", "))
	}
	return strings.Join(parts, "; ")
}

func isValidTeacherRole(role constants.TeacherRole) bool {
	return slices.Contains(constants.AllTeacherRoles(), role)
}

func roleSet(roles []constants.TeacherRole) map[constants.TeacherRole]struct{} {
	set := make(map[constants.TeacherRole]struct{}, len(roles))
	for _, role := range roles {
		set[role] = struct{}{}
	}
	return set
}
