package billing

import (
	"math"

	"zion-english/internal/constants"
)

const classRateHalfEpsilon = 0.01

// ClassRateIsHalfOfTeacherRate reports whether classRate is half of teacherRate (tbl_teachers.rate_per_class).
func ClassRateIsHalfOfTeacherRate(classRate, teacherRate float64) bool {
	if classRate <= 0 || teacherRate <= 0 {
		return false
	}
	half := teacherRate / 2
	return math.Abs(classRate-half) <= classRateHalfEpsilon
}

// EffectiveParentBillAmount returns the parent bill for one class row.
// Halving cancelled/rescheduled rows applies only when class rate is half of teacher rate.
// Keep SumConductedParentRateByCurrencyAndDateRange in class_record.sql in sync.
func EffectiveParentBillAmount(status constants.ClassStatus, parentRate, classRate, teacherRate float64) float64 {
	switch status {
	case constants.ClassStatusCancelled, constants.ClassStatusRescheduled:
		if ClassRateIsHalfOfTeacherRate(classRate, teacherRate) {
			return parentRate / 2
		}
		return parentRate
	default:
		return parentRate
	}
}
