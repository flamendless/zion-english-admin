package frontend

import (
	"zion-english/internal/constants"
	"zion-english/internal/teachers"
)

func WithRoleBadge(props AvatarProps, roles []constants.TeacherRole) AvatarProps {
	if role, ok := teachers.PrimaryTeacherRole(roles); ok {
		props.RoleBadge = string(role)
	}
	return props
}

func WithSuperuserBadge(props AvatarProps) AvatarProps {
	props.RoleBadge = "superuser"
	return props
}

func AvatarRoleBadgeTone(label string) PillTone {
	if label == "superuser" {
		return PillTonePrimary
	}
	if label == "" {
		return PillToneNeutral
	}
	return TeacherRolePillTone(constants.TeacherRole(label))
}
