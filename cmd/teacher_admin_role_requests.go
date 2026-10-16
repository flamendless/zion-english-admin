package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
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

func requireSuperuser(w http.ResponseWriter, r *http.Request) bool {
	if auth.GetRole(r.Context()) != auth.RoleSuperuser {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return false
	}
	return true
}

func handleTeacherAdminRoleRequests(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if !requireSuperuser(w, r) {
		return
	}

	ctx := r.Context()
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = string(constants.TeacherAdminRoleRequestStatusPending)
	}
	if !constants.ValidTeacherAdminRoleRequestStatus(status) {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}

	rows, err := dbRO.GetQueries().GetTeacherAdminRoleRequestsByStatus(ctx, status)
	if err != nil {
		logs.Log().Error("list admin role requests", zap.Error(err))
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}

	items := make([]frontend.AdminRoleRequestItem, len(rows))
	for i, row := range rows {
		targetName := utils.ComposePersonName(row.TargetFirstName, row.TargetMiddleName, row.TargetLastName)
		requesterName := utils.ComposePersonName(row.RequesterFirstName, row.RequesterMiddleName, row.RequesterLastName)
		items[i] = frontend.AdminRoleRequestItem{
			ID:          strconv.FormatInt(row.ID, 10),
			RequestedAt: utils.FormatPlanDateTimePHT(row.CreatedAt),
			TargetName: targetName,
			TargetAvatar: buildTeacherListAvatarProps(
				row.TargetTeacherID,
				row.TargetFirstName,
				row.TargetMiddleName,
				row.TargetLastName,
				row.TargetAssignedColor,
				row.TargetProfilePicture,
			),
			RequesterName: requesterName,
			RequesterAvatar: buildTeacherListAvatarProps(
				row.RequestedByTeacherID,
				row.RequesterFirstName,
				row.RequesterMiddleName,
				row.RequesterLastName,
				row.RequesterAssignedColor,
				row.RequesterProfilePicture,
			),
			CanReview: status == string(constants.TeacherAdminRoleRequestStatusPending),
		}
	}

	writeHTML(w)
	frontend.AdminRoleRequestsPage(frontend.AdminRoleRequestsPageData{
		Status: constants.TeacherAdminRoleRequestStatus(status),
		Items:  items,
	}).Render(ctx, w)
}

func handleTeacherAdminRoleRequestsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "admin-role-requests", "/approve"); ok {
		handleTeacherAdminRoleRequestReview(w, r, id, true)
		return
	}
	if id, ok := extractPathID(r, "admin-role-requests", "/reject"); ok {
		handleTeacherAdminRoleRequestReview(w, r, id, false)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleTeacherAdminRoleRequestReview(w http.ResponseWriter, r *http.Request, requestID int64, approve bool) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if !requireSuperuser(w, r) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	row, err := dbRO.GetQueries().GetTeacherAdminRoleRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			setErrorFlash(w, "Admin role request not found")
			HttpRedirect(w, r, "/admin-role-requests")
			return
		}
		setErrorFlash(w, MsgSomethingWrong)
		HttpRedirect(w, r, "/admin-role-requests")
		return
	}
	if row.Status != string(constants.TeacherAdminRoleRequestStatusPending) {
		setErrorFlash(w, "Only pending admin role requests can be reviewed")
		HttpRedirect(w, r, "/admin-role-requests")
		return
	}

	tx, err := dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		setErrorFlash(w, MsgFailedToStartTransaction)
		HttpRedirect(w, r, "/admin-role-requests")
		return
	}
	defer tx.Rollback()

	qtx := dbRW.GetQueries().WithTx(tx)

	newStatus := string(constants.TeacherAdminRoleRequestStatusRejected)
	actionLabel := "rejected"
	if approve {
		newStatus = string(constants.TeacherAdminRoleRequestStatusApproved)
		actionLabel = "approved"
		count, err := qtx.TeacherHasRole(ctx, queries.TeacherHasRoleParams{
			TeacherID: row.TargetTeacherID,
			Role:      string(constants.TeacherRoleAdmin),
		})
		if err != nil {
			setErrorFlash(w, MsgSomethingWrong)
			HttpRedirect(w, r, "/admin-role-requests")
			return
		}
		if count == 0 {
			if err := qtx.InsertTeacherRole(ctx, queries.InsertTeacherRoleParams{
				TeacherID: row.TargetTeacherID,
				Role:      string(constants.TeacherRoleAdmin),
			}); err != nil {
				setErrorFlash(w, MsgSomethingWrong)
				HttpRedirect(w, r, "/admin-role-requests")
				return
			}
		}
	}

	if err := qtx.UpdateTeacherAdminRoleRequestStatus(ctx, queries.UpdateTeacherAdminRoleRequestStatusParams{
		Status:     newStatus,
		ReviewedBy: sql.NullInt64{Int64: user.ID, Valid: true},
		ID:         requestID,
	}); err != nil {
		setErrorFlash(w, MsgSomethingWrong)
		HttpRedirect(w, r, "/admin-role-requests")
		return
	}

	if err := tx.Commit(); err != nil {
		setErrorFlash(w, MsgSomethingWrong)
		HttpRedirect(w, r, "/admin-role-requests")
		return
	}

	targetName := teacherNameByID(ctx, row.TargetTeacherID)
	requesterName := teacherNameByID(ctx, row.RequestedByTeacherID)
	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("%s admin role grant for teacher '%s' (request %d)", actionLabel, targetName, requestID))

	if approve {
		notifyTeacher(ctx, row.TargetTeacherID, targetName, user, notifications.KindProfileUpdated,
			"Your account was granted the admin role", "")
	}
	notifyTeacher(ctx, row.RequestedByTeacherID, requesterName, user, notifications.KindProfileUpdated,
		fmt.Sprintf("Admin role grant for '%s' was %s", targetName, actionLabel), "")

	setSuccessFlash(w, fmt.Sprintf("Admin role request %s successfully.", actionLabel))
	HttpRedirect(w, r, "/admin-role-requests")
}
