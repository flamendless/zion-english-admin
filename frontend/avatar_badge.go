package frontend

import (
	"zion-english/internal/constants"
	"zion-english/internal/models"
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

func AvatarForTimeline(props AvatarProps) AvatarProps {
	props.RoleBadge = ""
	if props.Size == "" {
		props.Size = "sm"
	}
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

func AvatarRoleBadgeToneForProps(props AvatarProps) PillTone {
	switch props.TeacherAccountStatus {
	case constants.TeacherStatusResigned:
		return PillToneError
	case constants.TeacherStatusPending:
		return PillToneWarning
	default:
		return AvatarRoleBadgeTone(props.RoleBadge)
	}
}

func AssignedTeacherAvatar(props AvatarProps, roles []constants.TeacherRole, status constants.TeacherStatus) AvatarProps {
	props = WithRoleBadge(props, roles)
	props.TeacherAccountStatus = status
	switch status {
	case constants.TeacherStatusResigned, constants.TeacherStatusPending:
		props.RoleBadge = string(status)
	}
	return props
}

func ApplyTeacherAvatarViewRoles(view models.AvatarView, roles []constants.TeacherRole, status constants.TeacherStatus) models.AvatarView {
	props := AssignedTeacherAvatar(AvatarProps{
		Initials:      view.Initials,
		AssignedColor: view.AssignedColor,
		PictureURL:    view.PictureURL,
		HasPicture:    view.HasPicture,
		Alt:           view.Alt,
		RoleBadge:     view.RoleBadge,
	}, roles, status)
	view.RoleBadge = props.RoleBadge
	if props.TeacherAccountStatus != "" {
		view.TeacherAccountStatus = string(props.TeacherAccountStatus)
	} else {
		view.TeacherAccountStatus = ""
	}
	return view
}
