package models

import "zion-english/internal/constants"

func ApplyTrialClassRate(req *ClassRecordRequest) {
	if req == nil || !req.IsTrialClass {
		return
	}
	req.Rate = constants.TrialClassRate
	req.Currency = constants.TrialClassCurrency
}

func ApplyScheduledTrialClassRate(req *ScheduledClassRequest) {
	if req == nil || !req.IsTrialClass {
		return
	}
	req.Rate = constants.TrialClassRate
	req.Currency = constants.TrialClassCurrency
}

func NormalizeClassRecordRate(req *ClassRecordRequest) {
	if req == nil {
		return
	}
	switch constants.ClassStatus(req.Status) {
	case constants.ClassStatusCancelled, constants.ClassStatusRescheduled:
		req.Rate = 0
	}
}
