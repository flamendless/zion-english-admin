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

