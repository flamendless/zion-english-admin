package frontend

import (
	"strings"

	"zion-english/internal/utils"
)

func RowClassForCutoffHalf(half utils.CutoffHalf) string {
	switch half {
	case utils.CutoffHalfFirst:
		return "row-cutoff-first"
	case utils.CutoffHalfSecond:
		return "row-cutoff-second"
	default:
		return ""
	}
}

func rowClassForCutoffDate(date string) string {
	half, ok := utils.CutoffHalfForDate(date)
	if !ok {
		return ""
	}
	return RowClassForCutoffHalf(half)
}

func rowClassForCutoffPeriod(startDate, endDate string) string {
	half, ok := utils.CutoffHalfForPeriod(startDate, endDate)
	if !ok {
		return ""
	}
	return RowClassForCutoffHalf(half)
}

func joinRowClasses(classes ...string) string {
	var out []string
	for _, c := range classes {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	return strings.Join(out, " ")
}

func ReportClassItemRowClass(date string) string {
	return rowClassForCutoffDate(date)
}
