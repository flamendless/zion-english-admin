package frontend

import "strings"
import "zion-english/internal/constants"

type PillTone string

const (
	PillToneInfo    PillTone = "info"
	PillToneSuccess PillTone = "success"
	PillToneWarning PillTone = "warning"
	PillToneError   PillTone = "error"
	PillToneNeutral PillTone = "neutral"
	PillTonePrimary PillTone = "primary"
)

type BadgeTone string

const (
	BadgeToneDestructive BadgeTone = "destructive"
	BadgeTonePrimary     BadgeTone = "primary"
)

func pillClass(tone PillTone) string {
	return "pill pill--" + string(tone)
}

func statusPillClass(tone PillTone) string {
	return "status-pill status-pill--" + string(tone)
}

func badgeCountClass(tone BadgeTone) string {
	return "badge-count badge-count--" + string(tone)
}

func StudentStatusPillTone(status string) PillTone {
	if status == "active" {
		return PillToneSuccess
	}
	return PillToneError
}

func TeacherStatusPillTone(status string, deleted bool) PillTone {
	if deleted {
		return PillToneNeutral
	}
	switch status {
	case "approved":
		return PillToneSuccess
	case "pending":
		return PillToneWarning
	default:
		return PillToneError
	}
}

func ClassStatusPillTone(status string) PillTone {
	switch status {
	case "conducted":
		return PillToneSuccess
	case "cancelled":
		return PillToneError
	case "rescheduled":
		return PillToneWarning
	case "scheduled":
		return PillTonePrimary
	default:
		return PillToneNeutral
	}
}

func DocumentStatusPillTone(status string) PillTone {
	switch status {
	case "submitted":
		return PillToneWarning
	case "approved":
		return PillToneSuccess
	case "rejected":
		return PillToneError
	default:
		return PillToneNeutral
	}
}

func AnnouncementLevelPillTone(level string) PillTone {
	switch level {
	case "info":
		return PillToneInfo
	case "warning":
		return PillToneWarning
	case "critical":
		return PillToneError
	default:
		return PillToneNeutral
	}
}

func AnnouncementSchedulePillTone(schedule string) PillTone {
	switch schedule {
	case "Active":
		return PillToneSuccess
	case "Upcoming":
		return PillToneWarning
	default:
		return PillToneNeutral
	}
}

func AnnouncementPublicationPillTone(status string) PillTone {
	switch status {
	case "published":
		return PillToneSuccess
	case "draft":
		return PillToneWarning
	default:
		return PillToneNeutral
	}
}

func NotificationReadPillTone(read bool) PillTone {
	if read {
		return PillToneNeutral
	}
	return PillTonePrimary
}

func RoleTagPillClass(role constants.TeacherRole) string {
	return "role-tag-pill " + pillClass(TeacherRolePillTone(role))
}

func TeacherRolePillTone(role constants.TeacherRole) PillTone {
	switch role {
	case constants.TeacherRoleAdmin:
		return PillTonePrimary
	case constants.TeacherRoleDeveloper:
		return PillToneInfo
	default:
		return PillToneNeutral
	}
}

func capitalizeStatus(status string) string {
	if status == "" {
		return ""
	}
	return strings.ToUpper(status[:1]) + status[1:]
}

func ratePillAriaLabel(rate float64, currency string) string {
	return formatRate(rate, currency)
}
