package payroll

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

func RecordTeacherStatusEvent(
	ctx context.Context,
	q *queries.Queries,
	sessionID, teacherID, changedByTeacherID int64,
	changedByName string,
	fromStatus string,
	toStatus constants.PayrollTeacherStatus,
	actionOrReason string,
) error {
	actorTeacherID, actorName, err := resolveStatusEventActor(ctx, q, sessionID, changedByTeacherID, changedByName)
	if err != nil {
		return err
	}
	var from sql.NullString
	if fromStatus != "" {
		from = sql.NullString{String: fromStatus, Valid: true}
	}
	var reasonNull sql.NullString
	if strings.TrimSpace(actionOrReason) != "" {
		reasonNull = sql.NullString{String: strings.TrimSpace(actionOrReason), Valid: true}
	}
	return q.InsertPayrollSessionTeacherStatusEvent(ctx, queries.InsertPayrollSessionTeacherStatusEventParams{
		SessionID:          sessionID,
		TeacherID:          teacherID,
		FromStatus:         from,
		ToStatus:           string(toStatus),
		Reason:             reasonNull,
		ChangedByTeacherID: actorTeacherID,
		ChangedByName:      actorName,
	})
}

func resolveStatusEventActor(ctx context.Context, q *queries.Queries, sessionID, changedByTeacherID int64, changedByName string) (int64, string, error) {
	actorID := changedByTeacherID
	if actorID == 0 {
		session, err := q.GetPayrollSessionByID(ctx, sessionID)
		if err != nil {
			return 0, "", err
		}
		actorID = session.StartedByTeacherID
	}
	name := strings.TrimSpace(changedByName)
	if name == "" {
		teacher, err := q.GetTeacherProfileByID(ctx, actorID)
		if err == nil {
			name = utils.ComposePersonName(teacher.FirstName, teacher.MiddleName, teacher.LastName)
		}
	}
	if name == "" {
		name = "Admin"
	}
	return actorID, name, nil
}

func OverrideTeacherStatus(
	ctx context.Context,
	q *queries.Queries,
	sessionID, teacherID, changedByTeacherID int64,
	changedByName string,
	newStatus constants.PayrollTeacherStatus,
	reason string,
) error {
	if reason == "" {
		return ErrOverrideReasonRequired
	}
	if !constants.ValidPayrollTeacherStatus(string(newStatus)) {
		return ErrInvalidTeacherStatus
	}
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
	if row.Status == string(newStatus) {
		return ErrStatusUnchanged
	}
	if newStatus != constants.PayrollTeacherStatusSentSalary && newStatus != constants.PayrollTeacherStatusSalaryReceived {
		if err := q.ClearPayrollSessionTeacherPayment(ctx, queries.ClearPayrollSessionTeacherPaymentParams{
			SessionID: sessionID,
			TeacherID: teacherID,
		}); err != nil {
			return err
		}
	}
	if err := q.SetPayrollSessionTeacherStatusDirect(ctx, queries.SetPayrollSessionTeacherStatusDirectParams{
		Status:    string(newStatus),
		SessionID: sessionID,
		TeacherID: teacherID,
	}); err != nil {
		return err
	}
	overrideReason := "Status override: " + reason
	return RecordTeacherStatusEvent(ctx, q, sessionID, teacherID, changedByTeacherID, changedByName, row.Status, newStatus, overrideReason)
}
