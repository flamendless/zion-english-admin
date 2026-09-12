package conf

import (
	"strings"

	"zion-english/internal/constants"
)

func MaskSecret(value string) string {
	switch {
	case value == "":
		return constants.StartupValueUnset
	case len(value) <= 2:
		return strings.Repeat("*", len(value))
	case len(value) == 3:
		return value[:1] + "**"
	case len(value) == 4:
		return value[:1] + "**" + value[3:]
	default:
		return value[:2] + strings.Repeat("*", len(value)-3) + value[len(value)-1:]
	}
}
