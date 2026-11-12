package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/payroll"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func handlePayrollPath(w http.ResponseWriter, r *http.Request) {
	sessionID, tail, ok := parsePayrollSessionsPath(r)
	if !ok {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	switch {
	case tail == "" && r.Method == http.MethodGet:
		handlePayrollSession(w, r, sessionID)
	case tail == "finish" && r.Method == http.MethodPost:
		handlePayrollSessionFinish(w, r, sessionID)
	case tail == "partials/teachers" && r.Method == http.MethodGet:
		handlePayrollSessionTeachersPartial(w, r, sessionID)
	default:
		if strings.HasPrefix(tail, "teachers/") && strings.HasSuffix(tail, "/edit") && r.Method == http.MethodGet {
			mid := strings.TrimSuffix(strings.TrimPrefix(tail, "teachers/"), "/edit")
			teacherID, err := strconv.ParseInt(mid, 10, 64)
			if err != nil || teacherID <= 0 {
				HttpError(w, MsgNotFound, http.StatusNotFound)
				return
			}
			handlePayrollTeacherEditModal(w, r, sessionID, teacherID)
			return
		}
		if strings.HasPrefix(tail, "teachers/") && strings.HasSuffix(tail, "/status") && r.Method == http.MethodPost {
			mid := strings.TrimSuffix(strings.TrimPrefix(tail, "teachers/"), "/status")
			teacherID, err := strconv.ParseInt(mid, 10, 64)
			if err != nil || teacherID <= 0 {
				HttpError(w, MsgNotFound, http.StatusNotFound)
				return
			}
			handlePayrollTeacherStatusOverride(w, r, sessionID, teacherID)
			return
		}
		if strings.HasPrefix(tail, "teachers/") && strings.HasSuffix(tail, "/audit") && r.Method == http.MethodPost {
			mid := strings.TrimSuffix(strings.TrimPrefix(tail, "teachers/"), "/audit")
			teacherID, err := strconv.ParseInt(mid, 10, 64)
			if err != nil || teacherID <= 0 {
				HttpError(w, MsgNotFound, http.StatusNotFound)
				return
			}
			handlePayrollTeacherAudit(w, r, sessionID, teacherID)
			return
		}
		if strings.HasPrefix(tail, "teachers/") && strings.HasSuffix(tail, "/generate") && r.Method == http.MethodPost {
			mid := strings.TrimSuffix(strings.TrimPrefix(tail, "teachers/"), "/generate")
			teacherID, err := strconv.ParseInt(mid, 10, 64)
			if err != nil || teacherID <= 0 {
				HttpError(w, MsgNotFound, http.StatusNotFound)
				return
			}
			handlePayrollTeacherGenerate(w, r, sessionID, teacherID)
			return
		}
		HttpError(w, MsgNotFound, http.StatusNotFound)
	}
}

func parsePayrollSessionsPath(r *http.Request) (sessionID int64, tail string, ok bool) {
	cfg := conf.Conf()
	prefix := strings.TrimSuffix(cfg.BasePath, "/") + "/payroll/sessions/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		return 0, "", false
	}
	rest := strings.TrimPrefix(r.URL.Path, prefix)
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, "", false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, "", false
	}
	if len(parts) == 1 {
		return id, "", true
	}
	return id, strings.Join(parts[1:], "/"), true
}

func handlePayroll(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	page := utils.ParsePageQuery(r)
	total, err := dbRO.GetQueries().CountPayrollSessions(ctx)
	if err != nil {
		HttpError(w, "Failed to load payroll sessions", http.StatusInternalServerError)
		return
	}
	page.Total = total
	rows, err := dbRO.GetQueries().ListPayrollSessions(ctx, queries.ListPayrollSessionsParams{
		Limit:  int64(page.Size),
		Offset: int64((page.Number - 1) * page.Size),
	})
	if err != nil {
		HttpError(w, "Failed to load payroll sessions", http.StatusInternalServerError)
		return
	}

	var activeID int64
	var activeLabel string
	if active, err := dbRO.GetQueries().GetActivePayrollSession(ctx); err == nil {
		activeID = active.ID
		activeLabel = payroll.CutoffLabel(active.PeriodStart, active.PeriodEnd)
	} else if !errors.Is(err, sql.ErrNoRows) {
		logs.Log().Warn("load active payroll session", zap.Error(err))
	}

	items := make([]frontend.PayrollSessionListItem, 0, len(rows))
	for _, row := range rows {
		startedBy := utils.ComposePersonName(row.StartedByFirstName, row.StartedByMiddleName, row.StartedByLastName)
		finishedAt := constants.DisplayNotSet
		if row.FinishedAt.Valid && row.FinishedAt.String != "" {
			finishedAt = formatPayrollTimestamp(row.FinishedAt.String)
		}
		items = append(items, frontend.PayrollSessionListItem{
			ID:             strconv.FormatInt(row.ID, 10),
			PeriodLabel:    payroll.CutoffLabel(row.PeriodStart, row.PeriodEnd),
			Status:         constants.PayrollSessionStatus(row.Status),
			RolesLabel:     frontend.ReportRolesFilterSummary(row.RoleTeacher == 1, row.RoleAdmin == 1),
			TeacherCount:   row.TeacherCount,
			ProgressLabel:  fmt.Sprintf("%d/%d", row.SalaryReceivedCount, row.TeacherCount),
			StartedAt:      formatPayrollTimestamp(row.StartedAt),
			FinishedAt:     finishedAt,
			StartedByName:  startedBy,
			IsActive:       row.Status == string(constants.PayrollSessionStatusActive),
		})
	}

	pag := frontend.BuildPaginationData(page.Number, page.Size, page.Total)
	prevURL := ""
	nextURL := ""
	if pag.HasPrev {
		prevURL = utils.URL(fmt.Sprintf("/payroll?page=%d", pag.PrevPage))
	}
	if pag.HasNext {
		nextURL = utils.URL(fmt.Sprintf("/payroll?page=%d", pag.NextPage))
	}

	writeHTML(w)
	if err := frontend.PayrollList(frontend.PayrollListData{
		Sessions:           items,
		Pagination:         pag,
		PrevURL:            prevURL,
		NextURL:            nextURL,
		HasActiveSession:   activeID > 0,
		ActiveSessionID:    strconv.FormatInt(activeID, 10),
		ActiveSessionLabel: activeLabel,
	}).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePayrollDatePresetPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	writeHTML(w)
	if err := frontend.DatePresetForMonthScoped(month, false, "payrollDatePresetGroup", "payrollDatePreset").Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePayrollStart(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid request")
		HttpRedirect(w, r, "/payroll")
		return
	}
	preset := strings.TrimSpace(r.FormValue("datePreset"))
	startDate, endDate := utils.CutoffDatesFromPreset(preset)
	if startDate == "" || endDate == "" {
		setErrorFlash(w, payroll.ErrInvalidPeriod.Error())
		HttpRedirect(w, r, "/payroll")
		return
	}
	filters := parseReportRoleFilters(r)
	ctx := r.Context()
	user := auth.GetUser(ctx)

	teacherIDs, err := payrollRosterTeacherIDs(ctx, startDate, endDate, filters)
	if err != nil {
		setErrorFlash(w, "Failed to build payroll roster")
		HttpRedirect(w, r, "/payroll")
		return
	}

	sessionID, err := payroll.StartSession(ctx, dbRW.GetQueries(), startDate, endDate, payroll.RoleFilters{
		Teacher: filters.Teacher,
		Admin:   filters.Admin,
	}, user.ID, teacherIDs)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/payroll")
		return
	}

	insertAuditLogAs(ctx, user, "payroll", fmt.Sprintf("Started payroll session #%d (%s to %s)", sessionID, startDate, endDate))
	setSuccessFlash(w, "Payroll session started")
	HttpRedirect(w, r, fmt.Sprintf("/payroll/sessions/%d", sessionID))
}

func handlePayrollSession(w http.ResponseWriter, r *http.Request, sessionID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	data, err := buildPayrollSessionData(ctx, sessionID)
	if err != nil {
		if errors.Is(err, payroll.ErrSessionNotFound) {
			HttpError(w, MsgNotFound, http.StatusNotFound)
			return
		}
		HttpError(w, "Failed to load payroll session", http.StatusInternalServerError)
		return
	}
	writeHTML(w)
	if err := frontend.PayrollSession(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePayrollSessionTeachersPartial(w http.ResponseWriter, r *http.Request, sessionID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writePayrollTeachersPartial(w, r.Context(), sessionID)
}

func writePayrollTeachersPartial(w http.ResponseWriter, ctx context.Context, sessionID int64) {
	session, err := dbRO.GetQueries().GetPayrollSessionByID(ctx, sessionID)
	if err != nil {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	rows, err := buildPayrollTeacherRows(ctx, session)
	if err != nil {
		HttpError(w, "Failed to load teachers", http.StatusInternalServerError)
		return
	}
	writeHTML(w)
	emptyMsg := "No teachers in this payroll session."
	if err := frontend.PayrollTeachersPartial(rows, emptyMsg, session.PeriodStart, session.PeriodEnd, session.Status == string(constants.PayrollSessionStatusActive)).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePayrollSessionFinish(w http.ResponseWriter, r *http.Request, sessionID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if err := payroll.FinishSession(ctx, dbRW.GetQueries(), sessionID); err != nil {
		if errors.Is(err, payroll.ErrFinishNotReady) {
			setErrorFlash(w, err.Error())
		} else {
			setErrorFlash(w, "Failed to finish payroll session")
		}
		HttpRedirect(w, r, fmt.Sprintf("/payroll/sessions/%d", sessionID))
		return
	}
	insertAuditLogAs(ctx, user, "payroll", fmt.Sprintf("Finished payroll session #%d", sessionID))
	setSuccessFlash(w, "Payroll session finished")
	HttpRedirect(w, r, "/payroll")
}

func handlePayrollTeacherAudit(w http.ResponseWriter, r *http.Request, sessionID, teacherID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if err := payroll.ConfirmAudit(ctx, dbRW.GetQueries(), sessionID, teacherID, user.ID, user.Name); err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	insertAuditLogAs(ctx, user, "payroll", fmt.Sprintf("Confirmed audit for teacher #%d in payroll session #%d", teacherID, sessionID))
	w.Header().Set(headerHXTrigger, `{"showSuccessBanner":"Audit confirmed"}`)
	writePayrollTeachersPartial(w, ctx, sessionID)
}

func handlePayrollTeacherGenerate(w http.ResponseWriter, r *http.Request, sessionID, teacherID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	session, err := dbRO.GetQueries().GetPayrollSessionByID(ctx, sessionID)
	if err != nil {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	if session.Status != string(constants.PayrollSessionStatusActive) {
		sendErrorLog(w, payroll.ErrSessionNotActive.Error())
		return
	}
	startDate := strings.TrimSpace(r.FormValue("startDate"))
	endDate := strings.TrimSpace(r.FormValue("endDate"))
	if startDate == "" || endDate == "" {
		startDate = session.PeriodStart
		endDate = session.PeriodEnd
	}
	if _, _, _, err := generateTeacherReportFile(ctx, teacherID, startDate, endDate); err != nil {
		sendErrorLog(w, "failed to generate report")
		return
	}
	rows, err := buildPayrollTeacherRows(ctx, session)
	if err != nil {
		sendErrorLog(w, "failed to load teacher row")
		return
	}
	var row frontend.PayrollTeacherRow
	for _, candidate := range rows {
		if candidate.TeacherID == strconv.FormatInt(teacherID, 10) {
			row = candidate
			break
		}
	}
	writeHTML(w)
	if err := frontend.PayrollTeacherTableRow(row, startDate, endDate, true, false).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePayrollDailyLogsConfirm(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(user.Role) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}
	row, err := dbRO.GetQueries().GetActivePayrollSessionTeacherRow(ctx, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			HttpError(w, MsgNotFound, http.StatusNotFound)
			return
		}
		sendErrorLog(w, "failed to load payroll session")
		return
	}
	if err := payroll.ConfirmDailyLogs(ctx, dbRW.GetQueries(), row.SessionID, user.ID, user.ID, user.Name); err != nil {
		sendErrorLog(w, "failed to confirm daily logs")
		return
	}
	insertAuditLogAs(ctx, user, "payroll", fmt.Sprintf("Confirmed daily logs for payroll session #%d", row.SessionID))
	if r.Header.Get(headerHXRequest) != "" {
		setHXRedirect(w, "/dashboard")
		return
	}
	HttpRedirect(w, r, "/dashboard")
}

func handlePayrollDailyLogsDefer(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(user.Role) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}
	row, err := dbRO.GetQueries().GetActivePayrollSessionTeacherRow(ctx, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	accessID, err := dbRO.GetQueries().GetLatestOpenAccessByTeacherID(ctx, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := dbRW.GetQueries().DeferPayrollDailyLogs(ctx, queries.DeferPayrollDailyLogsParams{
		DailyLogsDismissedAccessID: utils.NullInt64(accessID),
		SessionID:                  row.SessionID,
		TeacherID:                  user.ID,
	}); err != nil {
		sendErrorLog(w, "failed to defer daily logs prompt")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func payrollRosterTeacherIDs(ctx context.Context, startDate, endDate string, filters reportRoleFilters) ([]int64, error) {
	summaries, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, reportSearchParams("", startDate, endDate, filters))
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(summaries))
	for _, summary := range summaries {
		ids = append(ids, summary.TeacherID)
	}
	return ids, nil
}

func buildPayrollSessionData(ctx context.Context, sessionID int64) (frontend.PayrollSessionData, error) {
	session, err := dbRO.GetQueries().GetPayrollSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return frontend.PayrollSessionData{}, payroll.ErrSessionNotFound
		}
		return frontend.PayrollSessionData{}, err
	}
	teacherRows, err := buildPayrollTeacherRows(ctx, session)
	if err != nil {
		return frontend.PayrollSessionData{}, err
	}
	remaining, err := dbRO.GetQueries().CountPayrollSessionTeachersNotSalaryReceived(ctx, sessionID)
	if err != nil {
		return frontend.PayrollSessionData{}, err
	}
	canFinish := session.Status == string(constants.PayrollSessionStatusActive) && remaining == 0 && len(teacherRows) > 0
	finishBlocked := "All teachers must confirm salary received before finishing."
	if len(teacherRows) == 0 {
		finishBlocked = "No teachers in this session."
	}
	return frontend.PayrollSessionData{
		SessionID:           strconv.FormatInt(session.ID, 10),
		PeriodLabel:         payroll.CutoffLabel(session.PeriodStart, session.PeriodEnd),
		StartDate:           session.PeriodStart,
		EndDate:             session.PeriodEnd,
		Status:              constants.PayrollSessionStatus(session.Status),
		RoleTeacher:         session.RoleTeacher == 1,
		RoleAdmin:           session.RoleAdmin == 1,
		IsActive:            session.Status == string(constants.PayrollSessionStatusActive),
		CanFinish:           canFinish,
		FinishBlockedReason: finishBlocked,
		Teachers:            teacherRows,
	}, nil
}

func buildPayrollTeacherRows(ctx context.Context, session queries.TblPayrollSession) ([]frontend.PayrollTeacherRow, error) {
	pstRows, err := dbRO.GetQueries().ListPayrollSessionTeachers(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	isActive := session.Status == string(constants.PayrollSessionStatusActive)
	filters := reportRoleFilters{
		Teacher: session.RoleTeacher == 1,
		Admin:   session.RoleAdmin == 1,
	}
	classCounts, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, reportSearchParams("", session.PeriodStart, session.PeriodEnd, filters))
	if err != nil {
		return nil, err
	}
	classesByTeacher := map[int64]int64{}
	for _, row := range classCounts {
		classesByTeacher[row.TeacherID] = row.TotalClasses
	}

	teacherIDs := make([]int64, 0, len(pstRows))
	for _, row := range pstRows {
		teacherIDs = append(teacherIDs, row.TeacherID)
	}
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, err
	}
	statusMap, err := loadTeacherStatusesByIDs(ctx, teacherIDs)
	if err != nil {
		return nil, err
	}

	out := make([]frontend.PayrollTeacherRow, 0, len(pstRows))
	for _, row := range pstRows {
		name := utils.ComposePersonName(row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName)
		status := constants.PayrollTeacherStatus(row.Status)
		reportRow, _ := loadReportRow(ctx, row.TeacherID, session.PeriodStart, session.PeriodEnd)
		earningsHTML := reportRow.EarningsHTML()
		payrollRow := frontend.PayrollTeacherRow{
			TeacherID:       strconv.FormatInt(row.TeacherID, 10),
			TeacherName:     name,
			Avatar: avatarWithTeacherRoles(
				buildTeacherListAvatarProps(row.TeacherID, row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName, constants.DefaultTeacherAssignedColor, row.TeacherProfilePicture),
				rolesMap[row.TeacherID],
				statusMap[row.TeacherID],
			),
			Status:          status,
			TotalClasses:    classesByTeacher[row.TeacherID],
			EarningsHTML:    earningsHTML,
			SessionID:       strconv.FormatInt(session.ID, 10),
			StartDate:       session.PeriodStart,
			EndDate:         session.PeriodEnd,
			IsActiveSession: isActive,
		}
		payrollRow.ApplyActionGates(reportRow)
		out = append(out, payrollRow)
	}
	return out, nil
}

func handlePayrollTeacherEditModal(w http.ResponseWriter, r *http.Request, sessionID, teacherID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	data, err := buildPayrollTeacherEditModalData(ctx, sessionID, teacherID)
	if err != nil {
		if errors.Is(err, payroll.ErrTeacherNotInSession) || errors.Is(err, payroll.ErrSessionNotFound) {
			HttpError(w, MsgNotFound, http.StatusNotFound)
			return
		}
		sendErrorLog(w, "failed to load payroll teacher status")
		return
	}
	writeHTML(w)
	if err := frontend.PayrollTeacherEditModal(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePayrollTeacherStatusOverride(w http.ResponseWriter, r *http.Request, sessionID, teacherID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, "invalid form submission")
		return
	}
	statusRaw := strings.TrimSpace(r.FormValue("status"))
	reason := strings.TrimSpace(r.FormValue("reason"))
	if !constants.ValidPayrollTeacherStatus(statusRaw) {
		sendErrorLog(w, payroll.ErrInvalidTeacherStatus.Error())
		return
	}
	if reason == "" {
		sendErrorLog(w, payroll.ErrOverrideReasonRequired.Error())
		return
	}
	newStatus := constants.PayrollTeacherStatus(statusRaw)
	if err := payroll.OverrideTeacherStatus(ctx, dbRW.GetQueries(), sessionID, teacherID, user.ID, user.Name, newStatus, reason); err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	insertAuditLogAs(ctx, user, "payroll", fmt.Sprintf("Overrode payroll status for teacher #%d in session #%d to %s: %s", teacherID, sessionID, newStatus.Label(), reason))
	w.Header().Set(headerHXTrigger, `{"showSuccessBanner":"Payroll status updated"}`)
	writePayrollTeachersPartial(w, ctx, sessionID)
}

func buildPayrollTeacherEditModalData(ctx context.Context, sessionID, teacherID int64) (frontend.PayrollTeacherEditModalData, error) {
	if _, err := dbRO.GetQueries().GetPayrollSessionByID(ctx, sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return frontend.PayrollTeacherEditModalData{}, payroll.ErrSessionNotFound
		}
		return frontend.PayrollTeacherEditModalData{}, err
	}
	teacherRows, err := dbRO.GetQueries().ListPayrollSessionTeachers(ctx, sessionID)
	if err != nil {
		return frontend.PayrollTeacherEditModalData{}, err
	}
	var teacherName string
	var currentStatus constants.PayrollTeacherStatus
	found := false
	for _, row := range teacherRows {
		if row.TeacherID != teacherID {
			continue
		}
		teacherName = utils.ComposePersonName(row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName)
		currentStatus = constants.PayrollTeacherStatus(row.Status)
		found = true
		break
	}
	if !found {
		return frontend.PayrollTeacherEditModalData{}, payroll.ErrTeacherNotInSession
	}
	eventRows, err := dbRO.GetQueries().ListPayrollSessionTeacherStatusEvents(ctx, queries.ListPayrollSessionTeacherStatusEventsParams{
		SessionID: sessionID,
		TeacherID: teacherID,
	})
	if err != nil {
		return frontend.PayrollTeacherEditModalData{}, err
	}
	events := make([]frontend.PayrollTeacherStatusEventItem, 0, len(eventRows))
	for _, row := range eventRows {
		fromLabel := "Initial"
		if row.FromStatus.Valid && row.FromStatus.String != "" {
			fromLabel = constants.PayrollTeacherStatus(row.FromStatus.String).Label()
		}
		actionLabel := constants.DisplayNotSet
		if row.Reason.Valid && strings.TrimSpace(row.Reason.String) != "" {
			actionLabel = strings.TrimSpace(row.Reason.String)
		}
		changedBy := strings.TrimSpace(row.ChangedByName)
		if changedBy == "" {
			changedBy = utils.ComposePersonName(
				payrollEventNamePart(row.ChangedByFirstName),
				payrollEventNamePart(row.ChangedByMiddleName),
				payrollEventNamePart(row.ChangedByLastName),
			)
		}
		events = append(events, frontend.PayrollTeacherStatusEventItem{
			ChangedAtLabel: formatPayrollTimestamp(row.ChangedAt),
			FromLabel:      fromLabel,
			ToLabel:        constants.PayrollTeacherStatus(row.ToStatus).Label(),
			ActionLabel:    actionLabel,
			ChangedByName:  changedBy,
		})
	}
	return frontend.PayrollTeacherEditModalData{
		SessionID:     strconv.FormatInt(sessionID, 10),
		TeacherID:     strconv.FormatInt(teacherID, 10),
		TeacherName:   teacherName,
		CurrentStatus: currentStatus,
		Events:        events,
	}, nil
}

func formatPayrollTimestamp(raw string) string {
	return formatPaymentTimestamp(raw)
}

func payrollEventNamePart(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func populatePayrollDailyLogs(ctx context.Context, data *frontend.DashboardData, teacherID int64) {
	row, err := dbRO.GetQueries().GetActivePayrollSessionTeacherRow(ctx, teacherID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logs.Log().Warn("payroll daily logs lookup", zap.Error(err))
		}
		return
	}
	if row.Status != string(constants.PayrollTeacherStatusAccomplishDailyLogs) {
		return
	}
	data.ShowPayrollDailyLogs = true
	data.PayrollDailyLogs = frontend.PayrollDailyLogsData{
		SessionID:   strconv.FormatInt(row.SessionID, 10),
		PeriodLabel: payroll.CutoffLabel(row.PeriodStart, row.PeriodEnd),
	}
}
