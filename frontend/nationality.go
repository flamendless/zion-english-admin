package frontend

import "zion-english/internal/nationalities"

type NationalityOption struct {
	ID          string
	CountryName string
	Language    string
}

func NationalityISOSelected(options []NationalityOption, selected string) nationalities.ISO {
	for _, opt := range options {
		if opt.ID == selected {
			return nationalities.NormalizeISO(opt.ID)
		}
	}
	return ""
}
