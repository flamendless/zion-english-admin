package frontend

type NationalityOption struct {
	ID          string
	CountryName string
	Language    string
	FlagURL     string
}

func NationalityFlagForSelected(options []NationalityOption, selected string) string {
	for _, opt := range options {
		if opt.ID == selected {
			return opt.FlagURL
		}
	}
	return ""
}
