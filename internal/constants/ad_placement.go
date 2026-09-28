package constants

type AdPlacement string

const (
	AdPlacementTop          AdPlacement = "top"
	AdPlacementLeft         AdPlacement = "left"
	AdPlacementRight        AdPlacement = "right"
	AdPlacementBottom       AdPlacement = "bottom"
	AdPlacementTopAndBottom AdPlacement = "top_and_bottom"
	AdPlacementLeftAndRight AdPlacement = "left_and_right"
	AdPlacementAllSides     AdPlacement = "all_sides"
)

func ValidAdPlacement(placement string) bool {
	switch AdPlacement(placement) {
	case AdPlacementTop, AdPlacementLeft, AdPlacementRight, AdPlacementBottom,
		AdPlacementTopAndBottom, AdPlacementLeftAndRight, AdPlacementAllSides:
		return true
	default:
		return false
	}
}

func (p AdPlacement) Zones() []AdZone {
	switch p {
	case AdPlacementTop:
		return []AdZone{AdZoneTop}
	case AdPlacementLeft:
		return []AdZone{AdZoneLeft}
	case AdPlacementRight:
		return []AdZone{AdZoneRight}
	case AdPlacementBottom:
		return []AdZone{AdZoneBottom}
	case AdPlacementTopAndBottom:
		return []AdZone{AdZoneTop, AdZoneBottom}
	case AdPlacementLeftAndRight:
		return []AdZone{AdZoneLeft, AdZoneRight}
	case AdPlacementAllSides:
		return []AdZone{AdZoneTop, AdZoneLeft, AdZoneRight, AdZoneBottom}
	default:
		return nil
	}
}

type AdZone string

const (
	AdZoneTop    AdZone = "top"
	AdZoneLeft   AdZone = "left"
	AdZoneRight  AdZone = "right"
	AdZoneBottom AdZone = "bottom"
)

func ValidAdZone(zone string) bool {
	switch AdZone(zone) {
	case AdZoneTop, AdZoneLeft, AdZoneRight, AdZoneBottom:
		return true
	default:
		return false
	}
}
