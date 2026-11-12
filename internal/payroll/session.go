package payroll

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type RoleFilters struct {
	Teacher bool
	Admin   bool
}

func StartSession(
	ctx context.Context,
	q *queries.Queries,
	startDate, endDate string,
	filters RoleFilters,
	startedByTeacherID int64,
	rosterTeacherIDs []int64,
) (int64, error) {
	if startDate == "" || endDate == "" {
		return 0, ErrInvalidPeriod
	}
	if !filters.Teacher && !filters.Admin {
		return 0, ErrRoleFilterRequired
	}
	if len(rosterTeacherIDs) == 0 {
		return 0, ErrEmptyRoster
	}
	if _, err := q.GetActivePayrollSession(ctx); err == nil {
		return 0, ErrActiveSessionExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("get active payroll session: %w", err)
	}

	teacherFlag := int64(0)
	if filters.Teacher {
		teacherFlag = 1
	}
	adminFlag := int64(0)
	if filters.Admin {
		adminFlag = 1
	}

	sessionID, err := q.InsertPayrollSession(ctx, queries.InsertPayrollSessionParams{
		PeriodStart:        startDate,
		PeriodEnd:          endDate,
		RoleTeacher:        teacherFlag,
		RoleAdmin:          adminFlag,
		StartedByTeacherID: startedByTeacherID,
	})
	if err != nil {
		return 0, fmt.Errorf("insert payroll session: %w", err)
	}

	for _, teacherID := range rosterTeacherIDs {
		if err := q.InsertPayrollSessionTeacher(ctx, queries.InsertPayrollSessionTeacherParams{
			SessionID: sessionID,
			TeacherID: teacherID,
		}); err != nil {
			return 0, fmt.Errorf("insert payroll session teacher: %w", err)
		}
		if err := RecordTeacherStatusEvent(ctx, q, sessionID, teacherID, startedByTeacherID, "", "", constants.PayrollTeacherStatusAccomplishDailyLogs, StatusEventAddedToSession); err != nil {
			return 0, fmt.Errorf("record payroll teacher status: %w", err)
		}
	}
	return sessionID, nil
}

func FinishSession(ctx context.Context, q *queries.Queries, sessionID int64) error {
	session, err := q.GetPayrollSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSessionNotFound
		}
		return err
	}
	if session.Status != string(constants.PayrollSessionStatusActive) {
		return ErrSessionNotActive
	}
	remaining, err := q.CountPayrollSessionTeachersNotSalaryReceived(ctx, sessionID)
	if err != nil {
		return err
	}
	if remaining > 0 {
		return ErrFinishNotReady
	}
	return q.FinishPayrollSession(ctx, sessionID)
}

func ConfirmAudit(ctx context.Context, q *queries.Queries, sessionID, teacherID, changedByTeacherID int64, changedByName string) error {
	if err := requireActiveSessionTeacher(ctx, q, sessionID, teacherID); err != nil {
		return err
	}
	row, err := q.GetPayrollSessionTeacher(ctx, queries.GetPayrollSessionTeacherParams{
		SessionID: sessionID,
		TeacherID: teacherID,
	})
	if err != nil {
		return err
	}
	if err := q.UpdatePayrollSessionTeacherStatus(ctx, queries.UpdatePayrollSessionTeacherStatusParams{
		Status:    string(constants.PayrollTeacherStatusAudited),
		SessionID: sessionID,
		TeacherID: teacherID,
		Status_2:  string(constants.PayrollTeacherStatusDoneDailyLogs),
	}); err != nil {
		return err
	}
	return RecordTeacherStatusEvent(ctx, q, sessionID, teacherID, changedByTeacherID, changedByName, row.Status, constants.PayrollTeacherStatusAudited, StatusEventConfirmAudit)
}

func ConfirmDailyLogs(ctx context.Context, q *queries.Queries, sessionID, teacherID, changedByTeacherID int64, changedByName string) error {
	row, err := q.GetPayrollSessionTeacher(ctx, queries.GetPayrollSessionTeacherParams{
		SessionID: sessionID,
		TeacherID: teacherID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeacherNotInSession
		}
		return err
	}
	if err := q.ConfirmPayrollDailyLogs(ctx, queries.ConfirmPayrollDailyLogsParams{
		SessionID: sessionID,
		TeacherID: teacherID,
	}); err != nil {
		return err
	}
	return RecordTeacherStatusEvent(ctx, q, sessionID, teacherID, changedByTeacherID, changedByName, row.Status, constants.PayrollTeacherStatusDoneDailyLogs, StatusEventConfirmDailyLogs)
}

func CutoffLabel(startDate, endDate string) string {
	start, err := utils.ParseDatePHT(startDate)
	if err != nil || start == nil {
		return startDate + " - " + endDate
	}
	end, err := utils.ParseDatePHT(endDate)
	if err != nil || end == nil {
		return startDate + " - " + endDate
	}
	if start.Month() == end.Month() && start.Year() == end.Year() {
		return fmt.Sprintf("%s %d - %d", start.Format("January"), start.Day(), end.Day())
	}
	return utils.DatePHT(*start) + " - " + utils.DatePHT(*end)
}

func requireActiveSessionTeacher(ctx context.Context, q *queries.Queries, sessionID, teacherID int64) error {
	row, err := q.GetPayrollSessionTeacher(ctx, queries.GetPayrollSessionTeacherParams{
		SessionID: sessionID,
		TeacherID: teacherID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTeacherNotInSession
		}
		return err
	}
	if row.SessionStatus != string(constants.PayrollSessionStatusActive) {
		return ErrSessionNotActive
	}
	return nil
}

func OnPaymentSent(ctx context.Context, q *queries.Queries, teacherID int64, periodStart, periodEnd string, paymentID, changedByTeacherID int64, changedByName string) error {
	row, err := q.GetActivePayrollSessionTeacherForPeriod(ctx, queries.GetActivePayrollSessionTeacherForPeriodParams{
		TeacherID:   teacherID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if row.Status != string(constants.PayrollTeacherStatusAudited) {
		return nil
	}
	if err := q.SetPayrollSessionTeacherPayment(ctx, queries.SetPayrollSessionTeacherPaymentParams{
		PaymentID: sql.NullInt64{Int64: paymentID, Valid: true},
		SessionID: row.SessionID,
		TeacherID: teacherID,
	}); err != nil {
		return err
	}
	return RecordTeacherStatusEvent(ctx, q, row.SessionID, teacherID, changedByTeacherID, changedByName, row.Status, constants.PayrollTeacherStatusSentSalary, StatusEventSendSalary)
}

func OnPaymentReceived(ctx context.Context, q *queries.Queries, paymentID, changedByTeacherID int64, changedByName string) error {
	row, err := q.GetPayrollSessionTeacherByPaymentID(ctx, sql.NullInt64{Int64: paymentID, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	prevStatus := row.Status
	if err := q.SetPayrollSessionTeacherSalaryReceived(ctx, queries.SetPayrollSessionTeacherSalaryReceivedParams{
		SessionID: row.SessionID,
		TeacherID: row.TeacherID,
		PaymentID: sql.NullInt64{Int64: paymentID, Valid: true},
	}); err != nil {
		return err
	}
	return RecordTeacherStatusEvent(ctx, q, row.SessionID, row.TeacherID, changedByTeacherID, changedByName, prevStatus, constants.PayrollTeacherStatusSalaryReceived, StatusEventSalaryReceived)
}
