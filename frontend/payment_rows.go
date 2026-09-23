package frontend

import "strconv"

func paymentsTableColspan(showTeacherColumn bool) string {
	cols := 9
	if showTeacherColumn {
		cols++
	}
	return strconv.Itoa(cols)
}
