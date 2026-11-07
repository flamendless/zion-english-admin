package nationalities

import "strings"

// ISO is an ISO 3166-1 alpha-2 country code stored in tbl_nationalities.id.
type ISO string

func NormalizeISO(s string) ISO {
	return ISO(strings.ToLower(strings.TrimSpace(s)))
}

func (c ISO) Valid() bool {
	return c != "" && FlagSVG(string(c)) != ""
}
