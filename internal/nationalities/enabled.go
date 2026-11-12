package nationalities

import "slices"

var EnabledISO = []ISO{
	ISO("ph"), // Philippines
	ISO("kr"), // South Korea
	ISO("jp"), // Japan
	ISO("ca"), // Canada
}

func IsEnabledISO(iso ISO) bool {
	iso = NormalizeISO(string(iso))
	return slices.Contains(EnabledISO, iso)
}
