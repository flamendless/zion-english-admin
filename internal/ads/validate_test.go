package ads

import (
	"testing"

	"zion-english/internal/constants"
)

func TestValidateRequest(t *testing.T) {
	t.Parallel()
	base := Request{
		Name:                "Home top banner",
		Placement:           string(constants.AdPlacementTop),
		AdType:              string(constants.AdTypeAffiliate),
		Status:              string(constants.AdStatusDraft),
		RandomizeKind:       string(constants.AdRandomizePerPage),
		AffiliateProductIDs: []int64{1},
	}
	if err := ValidateRequest(base); err != nil {
		t.Fatalf("expected valid base: %v", err)
	}

	multi := base
	multi.AffiliateProductIDs = []int64{1, 2}
	multi.RandomizeKind = string(constants.AdRandomizeTimer)
	multi.TimerInterval = string(constants.AdTimerIntervalHourly)
	if err := ValidateRequest(multi); err != nil {
		t.Fatalf("expected valid multi timer: %v", err)
	}

	multiFiveSec := multi
	multiFiveSec.TimerInterval = string(constants.AdTimerIntervalFiveSeconds)
	if err := ValidateRequest(multiFiveSec); err != nil {
		t.Fatalf("expected valid multi 5 second timer: %v", err)
	}

	singleSession := base
	singleSession.RandomizeKind = string(constants.AdRandomizePerSession)
	normalizedSingle := NormalizeRequest(singleSession)
	if normalizedSingle.RandomizeKind != string(constants.AdRandomizePerPage) {
		t.Fatalf("expected per_page after normalize for single product, got %q", normalizedSingle.RandomizeKind)
	}
	if err := ValidateRequest(singleSession); err != nil {
		t.Fatalf("expected valid after normalize inside ValidateRequest: %v", err)
	}

	multiSession := multi
	multiSession.RandomizeKind = string(constants.AdRandomizePerSession)
	multiSession.TimerInterval = string(constants.AdTimerIntervalHourly)
	multiSession = NormalizeRequest(multiSession)
	if multiSession.TimerInterval != "" {
		t.Fatalf("expected timer interval cleared for non-timer kind, got %q", multiSession.TimerInterval)
	}
	if err := ValidateRequest(multiSession); err != nil {
		t.Fatalf("expected valid multi per_session: %v", err)
	}
}

func TestAdPlacementZones(t *testing.T) {
	t.Parallel()
	zones := constants.AdPlacementAllSides.Zones()
	if len(zones) != 4 {
		t.Fatalf("expected 4 zones, got %d", len(zones))
	}
	topBottom := constants.AdPlacementTopAndBottom.Zones()
	if len(topBottom) != 2 || topBottom[0] != constants.AdZoneTop || topBottom[1] != constants.AdZoneBottom {
		t.Fatalf("unexpected top and bottom zones: %v", topBottom)
	}
}
