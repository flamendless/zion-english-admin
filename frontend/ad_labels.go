package frontend

import (
	"zion-english/internal/constants"
)

func AdStatusLabel(status constants.AdStatus) string {
	switch status {
	case constants.AdStatusPublished:
		return "Published"
	case constants.AdStatusDraft:
		return "Draft"
	case constants.AdStatusDeleted:
		return "Deleted"
	default:
		return string(status)
	}
}

func AdStatusPillTone(status constants.AdStatus) PillTone {
	switch status {
	case constants.AdStatusPublished:
		return PillToneSuccess
	case constants.AdStatusDraft:
		return PillToneWarning
	case constants.AdStatusDeleted:
		return PillToneNeutral
	default:
		return PillToneNeutral
	}
}

func AdPlacementLabel(placement constants.AdPlacement) string {
	switch placement {
	case constants.AdPlacementTop:
		return "Top"
	case constants.AdPlacementLeft:
		return "Left"
	case constants.AdPlacementRight:
		return "Right"
	case constants.AdPlacementBottom:
		return "Bottom"
	case constants.AdPlacementTopAndBottom:
		return "Top and bottom"
	case constants.AdPlacementLeftAndRight:
		return "Left and right sides"
	case constants.AdPlacementAllSides:
		return "All sides"
	default:
		return string(placement)
	}
}

func AdZoneLabel(zone constants.AdZone) string {
	switch zone {
	case constants.AdZoneTop:
		return "Top"
	case constants.AdZoneLeft:
		return "Left"
	case constants.AdZoneRight:
		return "Right"
	case constants.AdZoneBottom:
		return "Bottom"
	default:
		return string(zone)
	}
}

func AdTypeLabel(adType constants.AdType) string {
	switch adType {
	case constants.AdTypeAffiliate:
		return "Affiliate"
	default:
		return string(adType)
	}
}

func AdRandomizeSummary(kind constants.AdRandomizeKind, interval constants.AdTimerInterval) string {
	switch kind {
	case constants.AdRandomizePerPage:
		return "Per page"
	case constants.AdRandomizePerSession:
		return "Per session"
	case constants.AdRandomizeTimer:
		switch interval {
		case constants.AdTimerIntervalFiveSeconds:
			return "Timer · 5 seconds"
		case constants.AdTimerIntervalHourly:
			return "Timer · Hourly"
		case constants.AdTimerIntervalDaily:
			return "Timer · Daily"
		default:
			return "Timer"
		}
	default:
		return string(kind)
	}
}
