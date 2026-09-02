package models

import "zion-english/internal/constants"

// NormalizeClassRecordRate forces rate to 0 for cancelled and rescheduled classes.
func NormalizeClassRecordRate(req *ClassRecordRequest) {
	if req == nil {
		return
	}
	switch constants.ClassStatus(req.Status) {
	case constants.ClassStatusCancelled, constants.ClassStatusRescheduled:
		req.Rate = 0
	}
}
