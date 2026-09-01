package frontend

import (
	"fmt"
	"zion-english/internal/utils"
)

func cutoffPresetForMonth(monthValue string, selectCurrentCutoff bool) (firstCutoff, secondCutoff, activeCutoff, normalizedMonth string) {
	year, month, ok := utils.ParseMonthPHT(monthValue)
	if !ok {
		normalizedMonth = utils.CurrentMonthPHT()
		year, month, _ = utils.ParseMonthPHT(normalizedMonth)
	} else {
		normalizedMonth = fmt.Sprintf("%04d-%02d", year, int(month))
	}

	firstCutoff, secondCutoff = utils.CutoffRangeForMonth(year, month)
	activeCutoff = firstCutoff
	if selectCurrentCutoff {
		activeCutoff = utils.ActiveCutoffForMonth(year, month)
	}
	return firstCutoff, secondCutoff, activeCutoff, normalizedMonth
}
