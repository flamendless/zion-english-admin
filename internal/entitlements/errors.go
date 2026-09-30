package entitlements

import "errors"

var (
	ErrUpgradeRequired           = errors.New("[ENTITLEMENTS] Teacher Pro is required for this feature")
	ErrAnalyticsRangeRequiresPro = errors.New("[ENTITLEMENTS] date range outside the current month requires Teacher Pro")
	ErrNoActiveTeacherPlanGrant  = errors.New("no active teacher plan grant")
)
