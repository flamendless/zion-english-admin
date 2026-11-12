package payroll

import "errors"

var (
	ErrActiveSessionExists     = errors.New("[PAYROLL] a payroll session is already active")
	ErrNoActiveSession         = errors.New("[PAYROLL] no active payroll session")
	ErrSessionNotFound         = errors.New("[PAYROLL] payroll session not found")
	ErrSessionNotActive        = errors.New("[PAYROLL] payroll session is not active")
	ErrInvalidPeriod           = errors.New("[PAYROLL] invalid payroll cutoff period")
	ErrRoleFilterRequired      = errors.New("[PAYROLL] select at least one role")
	ErrTeacherNotInSession     = errors.New("[PAYROLL] teacher is not in this payroll session")
	ErrInvalidStatusTransition = errors.New("[PAYROLL] invalid payroll status for this action")
	ErrFinishNotReady          = errors.New("[PAYROLL] all teachers must confirm salary received before finishing")
	ErrEmptyRoster             = errors.New("[PAYROLL] no teachers match the selected period and roles")
	ErrOverrideReasonRequired  = errors.New("[PAYROLL] override reason is required")
	ErrInvalidTeacherStatus    = errors.New("[PAYROLL] invalid payroll teacher status")
	ErrStatusUnchanged         = errors.New("[PAYROLL] choose a different status to override")
)
