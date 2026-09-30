package frontend

import (
	"strings"

	"zion-english/internal/constants"
)

func PlanTierPillTone(tierLabel string) PillTone {
	if strings.EqualFold(strings.TrimSpace(tierLabel), "Teacher Pro") {
		return PillToneSuccess
	}
	return PillToneNeutral
}

func PlanTransactionStatusPillTone(status string) PillTone {
	switch strings.TrimSpace(status) {
	case "Active":
		return PillToneSuccess
	case "Scheduled":
		return PillToneWarning
	case "Expired":
		return PillToneNeutral
	case "Revoked":
		return PillToneError
	default:
		return PillToneNeutral
	}
}

func TeacherPlanBillingLabel(kind constants.TeacherPlanBillingKind) string {
	switch kind {
	case constants.TeacherPlanBillingMonthly:
		return "Monthly"
	case constants.TeacherPlanBillingLifetime:
		return "Lifetime"
	default:
		return string(kind)
	}
}

func TeacherPlanBillingPillTone(kind constants.TeacherPlanBillingKind) PillTone {
	switch kind {
	case constants.TeacherPlanBillingLifetime:
		return PillTonePrimary
	case constants.TeacherPlanBillingMonthly:
		return PillToneInfo
	default:
		return PillToneNeutral
	}
}
