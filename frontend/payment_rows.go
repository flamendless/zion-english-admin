package frontend

import "strconv"

func paymentsTableColspan(showTeacherColumn bool) string {
	return strconv.Itoa(paymentsTableColspanInt(showTeacherColumn))
}

func paymentsTableColspanInt(showTeacherColumn bool) int {
	cols := 9
	if showTeacherColumn {
		cols++
	}
	return cols
}

func (r PaymentRowData) RowClass() string {
	return rowClassForCutoffPeriod(r.PeriodStart, r.PeriodEnd)
}
