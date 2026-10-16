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
// Conducted: scales parent_rate by class duration vs teacher rate_duration_minutes.
// Cancelled/rescheduled: class rate 0 bills 0 parent; half teacher rate bills half parent; else full parent.
// Keep SumConductedParentRateByCurrencyAndDateRange in class_record.sql in sync.
func EffectiveParentBillAmount(
	status constants.ClassStatus,
	parentRate, classRate, teacherRate float64,
	classDurationMinutes, teacherRateDurationMinutes int64,
) float64 {
	switch status {
	case constants.ClassStatusConducted:
		if classDurationMinutes > 0 && teacherRateDurationMinutes > 0 {
			return parentRate * float64(classDurationMinutes) / float64(teacherRateDurationMinutes)
		}
		return parentRate
	case constants.ClassStatusCancelled, constants.ClassStatusRescheduled:
		if classRate == 0 {
			return 0
		}
		if ClassRateIsHalfOfTeacherRate(classRate, teacherRate) {
			return parentRate / 2
		}
		return parentRate
	default:
		return parentRate
	}
}
