package frontend

import (
	"strings"

	"zion-english/internal/utils"
)

type EarningsCutoffRow struct {
	MonthLabel  string
	CutoffKind  EarningsCutoffKind
	PeriodLabel string
	Earnings    []CurrencyTotal
	Conducted   int64
}

func (r EarningsCutoffRow) EarningsHTML() string {
	if len(r.Earnings) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(r.Earnings))
	for _, item := range r.Earnings {
		parts = append(parts, utils.FormatCurrency(item.Total, item.Currency))
	}
	return strings.Join(parts, "<br>")
}
