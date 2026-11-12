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
	const open = "<svg"
	before, after, ok := strings.Cut(svg, open)
	if !ok {
		return svg
	}
	rest := strings.TrimSpace(after)
	if strings.HasPrefix(rest, "class=") {
		return svg
	}
	return before + `<svg class="` + class + `" ` + rest
}
