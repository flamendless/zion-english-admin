package frontend

import (
	"strings"

	"zion-english/internal/nationalities"
)

func flagMarkup(iso, class string) string {
	svg := nationalities.FlagSVG(iso)
	if svg == "" {
		return ""
	}
	if class == "" {
		return svg
	}
	const open = "<svg "
	if !strings.HasPrefix(svg, open) {
		return svg
	}
	return strings.Replace(svg, open, `<svg class="`+class+`" `, 1)
}
