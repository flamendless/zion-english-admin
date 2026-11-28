package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/notifications"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func validateLeaveRequestForm(leaveType, startDate, endDate, reason string) error {
	if !constants.ValidLeaveRequestType(leaveType) {
		return ErrInvalidLeaveRequestType
	}
	startNorm := utils.NormalizeDatePHT(startDate)
	if startNorm == "" {
		return ErrInvalidStartDateFormat
	}
	endNorm := utils.NormalizeDatePHT(endDate)
	if endNorm == "" {
		return ErrInvalidEndDateFormat
	}
	if endNorm < startNorm {
		return ErrLeaveRequestEndBeforeStart
	}
	if strings.TrimSpace(reason) == "" {
		return ErrLeaveRequestReasonRequired
	}
	return nil
}

func mapLeaveRequestListRow(
	row queries.ListTeacherLeaveRequestsRow,
	showTeacher bool,
) frontend.LeaveRequestItem {
	name := utils.ComposePersonName(row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName)
	return frontend.LeaveRequestItem{
		ID:          strconv.FormatInt(row.ID, 10),
		TeacherName: name,
		TeacherAvatar: buildTeacherListAvatarProps(
			row.TeacherID,
			row.TeacherFirstName,
			row.TeacherMiddleName,
			row.TeacherLastName,
			row.TeacherAssignedColor,
			row.TeacherProfilePicture,
		),
		ShowTeacher: showTeacher,
		LeaveType:   constants.LeaveRequestType(row.LeaveType),
		StartDate:   utils.NormalizeDatePHT(row.StartDate),
		EndDate:     utils.NormalizeDatePHT(row.EndDate),
		Reason:      row.Reason,
		Status:      constants.LeaveRequestStatus(row.Status),
		SubmittedAt: utils.FormatPlanDateTimePHT(row.CreatedAt),
		CanReview:   row.Status == string(constants.LeaveRequestStatusSubmitted),
	}
}

func mapLeaveRequestTeacherRow(row queries.GetTeacherLeaveRequestsRow) frontend.LeaveRequestItem {
	return frontend.LeaveRequestItem{
		ID:          strconv.FormatInt(row.ID, 10),
		ShowTeacher: false,
		LeaveType:   constants.LeaveRequestType(row.LeaveType),
		StartDate:   utils.NormalizeDatePHT(row.StartDate),
		EndDate:     utils.NormalizeDatePHT(row.EndDate),
		Reason:      row.Reason,
		Status:      constants.LeaveRequestStatus(row.Status),
		SubmittedAt: utils.FormatPlanDateTimePHT(row.CreatedAt),
		CanReview:   false,
	}
}

func handleLeaveRequests(w http.ResponseWriter, r *http.Request) {
	role := auth.GetRole(r.Context())

	switch r.Method {
	case http.MethodGet:
		handleLeaveRequestsList(w, r)
	case http.MethodPost:
		if role != auth.RoleTeacher {
			HttpError(w, MsgAccessDenied, http.StatusForbidden)
			return
		}
		handleLeaveRequestCreate(w, r)
	default:
		requireMethod(w, r, http.MethodGet)
	}
}

func handleLeaveRequestsList(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)

	data := frontend.LeaveRequestsPageData{
		TodayPHT: utils.TodayPHT(),
	}

	switch role {
	case auth.RoleSuperuser, auth.RoleAdmin:
		data.Title = "Time-off requests"
		data.Description = "Review teacher time-off submissions."
		data.ShowTeacherColumn = true
		data.ShowTeacherFilter = true
		data.CanRequestLeave = false
	case auth.RoleTeacher:
		data.Title = "Time-off requests"
		data.Description = "Submit and track your time-off requests."
		data.CanRequestLeave = true
	default:
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && !constants.ValidLeaveRequestStatus(status) {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	data.Status = constants.LeaveRequestStatus(status)

	var items []frontend.LeaveRequestItem
	switch role {
	case auth.RoleTeacher:
		rows, err := dbRO.GetQueries().GetTeacherLeaveRequests(ctx, queries.GetTeacherLeaveRequestsParams{
			TeacherID: user.ID,
			Column2:   status,
			Status:    status,
		})
		if err != nil {
			logs.Log().Error("list teacher leave requests", zap.Error(err))
			HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
			return
		}
		items = make([]frontend.LeaveRequestItem, len(rows))
		for i, row := range rows {
			items[i] = mapLeaveRequestTeacherRow(row)
		}
	case auth.RoleSuperuser, auth.RoleAdmin:
		teacherID := int64(0)
		if tid := strings.TrimSpace(r.URL.Query().Get("teacherId")); tid != "" {
			parsed, err := strconv.ParseInt(tid, 10, 64)
			if err != nil || parsed <= 0 {
				HttpError(w, ErrInvalidTeacherID.Error(), http.StatusBadRequest)
				return
			}
			teacherID = parsed
		}
		data.TeacherFilterID = strconv.FormatInt(teacherID, 10)
		if teacherID > 0 {
			data.TeacherFilterName = teacherNameByID(ctx, teacherID)
		}
		rows, err := dbRO.GetQueries().ListTeacherLeaveRequests(ctx, queries.ListTeacherLeaveRequestsParams{
			Column1:   status,
			Status:    status,
			Column3:   teacherID,
			TeacherID: teacherID,
		})
		if err != nil {
			logs.Log().Error("list leave requests", zap.Error(err))
			HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
			return
		}
		items = make([]frontend.LeaveRequestItem, len(rows))
		for i, row := range rows {
			items[i] = mapLeaveRequestListRow(row, true)
		}
	}

	data.Items = items
	if err := frontend.LeaveRequestsPage(data).Render(ctx, w); err != nil {
		logs.Log().Error("render leave requests page", zap.Error(err))
	}
}

func handleLeaveRequestCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form submission")
		HttpRedirect(w, r, "/leave-requests")
		return
	}

	leaveType := strings.TrimSpace(r.FormValue("leave_type"))
	startDate := strings.TrimSpace(r.FormValue("start_date"))
	endDate := strings.TrimSpace(r.FormValue("end_date"))
	reason := strings.TrimSpace(r.FormValue("reason"))

	if err := validateLeaveRequestForm(leaveType, startDate, endDate, reason); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/leave-requests")
		return
	}

	startNorm := utils.NormalizeDatePHT(startDate)
	endNorm := utils.NormalizeDatePHT(endDate)

	err := dbRW.GetQueries().InsertTeacherLeaveRequest(ctx, queries.InsertTeacherLeaveRequestParams{
		TeacherID: user.ID,
		LeaveType: leaveType,
		StartDate: startNorm,
		EndDate:   endNorm,
		Reason:    reason,
	})
	if err != nil {
		logs.Log().Error("insert leave request", zap.Error(err))
		setErrorFlash(w, MsgSomethingWrong)
		HttpRedirect(w, r, "/leave-requests")
		return
	}

	typeLabel := constants.LeaveRequestType(leaveType).Label()
	insertAuditLogAs(ctx, user, "leave-requests", fmt.Sprintf(
		"submitted %s time-off %s to %s",
		typeLabel,
		startNorm,
		endNorm,
	))
	notifySuperuser(ctx, user, notifications.KindLeaveRequestSubmitted,
		fmt.Sprintf("Teacher '%s' submitted %s time-off (%s to %s)", user.Name, typeLabel, startNorm, endNorm), "")

	setSuccessFlash(w, "Time-off request submitted successfully. An administrator will review it.")
	redirectPath := "/leave-requests"
	if listStatus := strings.TrimSpace(r.FormValue("list_status")); listStatus != "" {
		if constants.ValidLeaveRequestStatus(listStatus) {
			redirectPath = "/leave-requests?status=" + url.QueryEscape(listStatus)
		}
	}
	HttpRedirect(w, r, redirectPath)
}

func handleLeaveRequestsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "leave-requests", "/approve"); ok {
		handleLeaveRequestReview(w, r, id, true)
		return
	}
	if id, ok := extractPathID(r, "leave-requests", "/reject"); ok {
		handleLeaveRequestReview(w, r, id, false)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleLeaveRequestReview(w http.ResponseWriter, r *http.Request, requestID int64, approve bool) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !auth.HasAdminAccess(auth.GetRole(r.Context())) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	row, err := dbRO.GetQueries().GetTeacherLeaveRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			setErrorFlash(w, "Time-off request not found")
			HttpRedirect(w, r, "/leave-requests")
			return
		}
		setErrorFlash(w, MsgSomethingWrong)
		HttpRedirect(w, r, "/leave-requests")
		return
	}
	if row.Status != string(constants.LeaveRequestStatusSubmitted) {
		setErrorFlash(w, "Only submitted time-off requests can be reviewed")
		HttpRedirect(w, r, "/leave-requests")
		return
	}

	newStatus := string(constants.LeaveRequestStatusRejected)
	actionLabel := "rejected"
	if approve {
		newStatus = string(constants.LeaveRequestStatusApproved)
		actionLabel = "approved"
	}

	if err := dbRW.GetQueries().UpdateTeacherLeaveRequestStatus(ctx, queries.UpdateTeacherLeaveRequestStatusParams{
		Status:     newStatus,
		ReviewedBy: sql.NullInt64{Int64: user.ID, Valid: true},
		ID:         requestID,
	}); err != nil {
		logs.Log().Error("update leave request status", zap.Error(err))
		setErrorFlash(w, MsgSomethingWrong)
		HttpRedirect(w, r, "/leave-requests")
		return
	}

	teacherName := teacherNameByID(ctx, row.TeacherID)
	typeLabel := constants.LeaveRequestType(row.LeaveType).Label()
	insertAuditLogAs(ctx, user, "leave-requests", fmt.Sprintf(
		"%s %s time-off request %d for teacher '%s' (%s to %s)",
		actionLabel,
		typeLabel,
		requestID,
		teacherName,
		row.StartDate,
		row.EndDate,
	))

	notifyTeacher(ctx, row.TeacherID, teacherName, user, notifications.KindLeaveRequestReviewed,
		fmt.Sprintf("Your %s time-off request (%s to %s) was %s", typeLabel, row.StartDate, row.EndDate, actionLabel), "")

	setSuccessFlash(w, fmt.Sprintf("Time-off request %s successfully.", actionLabel))
	HttpRedirect(w, r, "/leave-requests?status=submitted")
}
