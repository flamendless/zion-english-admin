package announcements

import (
	"strings"

	"zion-english/internal/utils"
)

func ResolveCTAURL(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return utils.URL(raw)
}

func IsExternalCTAURL(raw string) bool {
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}
