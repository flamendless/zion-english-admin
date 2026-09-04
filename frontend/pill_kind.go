package frontend

import (
	"strings"

	"zion-english/internal/constants"
)

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

func StudentStatusPillTone(status constants.StudentStatus) PillTone {
	if status == constants.StudentStatusActive {
		return PillToneSuccess
	}
	return PillToneError
}

func TeacherStatusPillTone(status constants.TeacherStatus, deleted bool) PillTone {
	if deleted {
		return PillToneNeutral
	}
	switch status {
	case constants.TeacherStatusApproved:
		return PillToneSuccess
	case constants.TeacherStatusPending:
		return PillToneWarning
	default:
		return PillToneError
	}
}

func ClassStatusPillTone(status constants.ClassListFilterStatus) PillTone {
	switch status {
	case constants.ClassListFilterConducted:
		return PillToneSuccess
	case constants.ClassListFilterCancelled:
		return PillToneError
	case constants.ClassListFilterRescheduled:
		return PillToneWarning
	case constants.ClassListFilterScheduled:
		return PillTonePrimary
	default:
		return PillToneNeutral
	}
}

func ClassRecordStatusPillTone(status constants.ClassStatus) PillTone {
	return ClassStatusPillTone(constants.ClassListFilterStatus(status))
}

func DocumentStatusPillTone(status constants.TeacherDocumentStatus) PillTone {
	switch status {
	case constants.TeacherDocumentStatusSubmitted:
		return PillToneWarning
	case constants.TeacherDocumentStatusApproved:
		return PillToneSuccess
	case constants.TeacherDocumentStatusRejected:
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
	case constants.TeacherRoleTester:
		return PillToneWarning
	default:
		return PillToneNeutral
	}
}

func capitalizeStatus[S ~string](status S) string {
	if status == "" {
		return ""
	}
	s := string(status)
	return strings.ToUpper(s[:1]) + s[1:]
}

func ratePillAriaLabel(rate float64, currency string) string {
	return formatRate(rate, currency)
}
