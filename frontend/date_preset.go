package frontend

import (
	"fmt"
	"zion-english/internal/utils"
)

func cutoffPresetForMonth(monthValue string, selectCurrentCutoff bool) (firstCutoff, secondCutoff, monthCutoff, defaultCutoff, normalizedMonth string) {
	year, month, ok := utils.ParseMonthPHT(monthValue)
	if !ok {
		normalizedMonth = utils.CurrentMonthPHT()
		year, month, _ = utils.ParseMonthPHT(normalizedMonth)
	} else {
		normalizedMonth = fmt.Sprintf("%04d-%02d", year, int(month))
	}

	firstCutoff, secondCutoff = utils.CutoffRangeForMonth(year, month)
	monthCutoff = utils.MonthCutoffPreset(year, month)
	defaultCutoff = firstCutoff
	if selectCurrentCutoff {
		defaultCutoff = monthCutoff
	}
	return firstCutoff, secondCutoff, monthCutoff, defaultCutoff, normalizedMonth
}
