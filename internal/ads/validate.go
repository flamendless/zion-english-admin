package ads

import (
	"strings"

	"zion-english/internal/constants"
)

type Request struct {
	Name                string
	Placement           string
	AdType              string
	Status              string
	RandomizeKind       string
	TimerInterval       string
	SortOrder           int64
	AffiliateProductIDs []int64
}

func NormalizeRequest(req Request) Request {
	req.Name = strings.TrimSpace(req.Name)
	req.Placement = strings.TrimSpace(req.Placement)
	req.AdType = strings.TrimSpace(req.AdType)
	req.Status = strings.TrimSpace(req.Status)
	req.RandomizeKind = strings.TrimSpace(req.RandomizeKind)
	req.TimerInterval = strings.TrimSpace(req.TimerInterval)
	if req.AdType == "" {
		req.AdType = string(constants.AdTypeAffiliate)
	}
	if req.RandomizeKind == "" {
		req.RandomizeKind = string(constants.AdRandomizePerPage)
	}
	if req.Status == "" {
		req.Status = string(constants.AdStatusDraft)
	}
	if len(req.AffiliateProductIDs) <= 1 {
		req.RandomizeKind = string(constants.AdRandomizePerPage)
		req.TimerInterval = ""
	} else if req.RandomizeKind != string(constants.AdRandomizeTimer) {
		req.TimerInterval = ""
	}
	return req
}

func ValidateRequest(req Request) error {
	req = NormalizeRequest(req)
	if req.Name == "" {
		return ErrNameRequired
	}
	if !constants.ValidAdPlacement(req.Placement) {
		return ErrInvalidPlacement
	}
	if !constants.ValidAdType(req.AdType) {
		return ErrInvalidAdType
	}
	if !constants.ValidAdFormStatus(req.Status) {
		return ErrInvalidFormStatus
	}
	if !constants.ValidAdRandomizeKind(req.RandomizeKind) {
		return ErrInvalidRandomizeKind
	}
	if req.AdType == string(constants.AdTypeAffiliate) && len(req.AffiliateProductIDs) == 0 {
		return ErrAffiliateProductsRequired
	}
	productCount := len(req.AffiliateProductIDs)
	if productCount <= 1 {
		if req.RandomizeKind != string(constants.AdRandomizePerPage) {
			return ErrTooManyProductsForKind
		}
		if req.TimerInterval != "" {
			return ErrTimerIntervalNotAllowed
		}
	} else {
		if req.RandomizeKind == string(constants.AdRandomizePerPage) ||
			req.RandomizeKind == string(constants.AdRandomizePerSession) ||
			req.RandomizeKind == string(constants.AdRandomizeTimer) {
			// ok
		} else {
			return ErrRandomizeKindRequired
		}
		if req.RandomizeKind == string(constants.AdRandomizeTimer) {
			if !constants.ValidAdTimerInterval(req.TimerInterval) || req.TimerInterval == "" {
				return ErrInvalidTimerInterval
			}
		} else if req.TimerInterval != "" {
			return ErrTimerIntervalNotAllowed
		}
	}
	return nil
}

func ValidateStatusUpdate(status string) error {
	if !constants.ValidAdStatus(status) {
		return ErrInvalidStatus
	}
	return nil
}
