package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/entitlements"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

const plansRecentTransactionLimit = 100

func handlePlans(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	teachers, err := buildPlansTeacherRows(ctx)
	if err != nil {
		logs.Log().Error("plans teachers", zap.Error(err))
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	transactions, err := listAllRecentPlanTransactions(ctx)
	if err != nil {
		logs.Log().Error("plans transactions", zap.Error(err))
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	data := frontend.PlansPageData{
		Teachers:     teachers,
		Transactions: transactions,
	}
	writeHTML(w)
	if err := frontend.PlansPage(data).Render(ctx, w); err != nil {
		logs.Log().Error("render plans", zap.Error(err))
	}
}

func handlePlansTeacherPath(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if id, ok := extractPathID(r, "plans/teacher", "/view"); ok {
		handlePlansTeacherView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "plans/teacher", "/grant"); ok {
		handlePlansTeacherGrantModal(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handlePlansTeacherView(w http.ResponseWriter, r *http.Request, teacherID int64) {
	ctx := r.Context()
	data, err := buildPlansTeacherViewModal(ctx, teacherID)
	if err != nil {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	writeHTML(w)
	if err := frontend.PlansTeacherViewModal(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePlansTeacherGrantModal(w http.ResponseWriter, r *http.Request, teacherID int64) {
	ctx := r.Context()
	data, err := buildPlansGrantModal(ctx, teacherID)
	if err != nil {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	writeHTML(w)
	if err := frontend.PlansGrantModal(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePlansGrant(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, MsgSomethingWrong)
		return
	}
	teacherID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("teacher_id")), 10, 64)
	if err != nil || teacherID <= 0 {
		sendErrorLog(w, ErrInvalidTeacherID.Error())
		return
	}
	billingKind := constants.TeacherPlanBillingKind(strings.TrimSpace(r.FormValue("billing_kind")))
	if billingKind != constants.TeacherPlanBillingMonthly && billingKind != constants.TeacherPlanBillingLifetime {
		sendErrorLog(w, "invalid billing kind")
		return
	}
	startDate := utils.NormalizeDatePHT(r.FormValue("effective_start"))
	if startDate == "" {
		startDate = utils.TodayPHT()
	}
	effectiveStart := startDate + " 00:00:00"
	var effectiveEnd sql.NullString
	if billingKind == constants.TeacherPlanBillingMonthly {
		end, err := database.MonthlyPlanEffectiveEndPHT(startDate)
		if err != nil {
			sendErrorLog(w, ErrInvalidStartDateFormat.Error())
			return
		}
		effectiveEnd = utils.NullIfEmptyString(end)
	}
	actor := auth.GetUser(ctx)
	row, err := dbRW.GetQueries().InsertTeacherPlanTransaction(ctx, queries.InsertTeacherPlanTransactionParams{
		TeacherID:          teacherID,
		Tier:               string(constants.TeacherPlanTierPro),
		BillingKind:        string(billingKind),
		EffectiveStart:     effectiveStart,
		EffectiveEnd:       effectiveEnd,
		GrantedByTeacherID: utils.NullInt64(actor.ID),
		GrantedByName:      actor.Name,
		Note:               strings.TrimSpace(r.FormValue("note")),
	})
	if err != nil {
		logs.Log().Error("insert plan transaction", zap.Error(err))
		sendErrorLog(w, MsgSomethingWrong)
		return
	}
	insertAuditLogAs(ctx, actor, "plans", fmt.Sprintf(
		"granted Teacher Pro (%s) to teacher #%d (transaction #%d)",
		billingKind, teacherID, row.ID,
	))
	setSuccessFlash(w, "Pro plan granted.")
	HttpRedirect(w, r, "/plans")
}

func handlePlansRevoke(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, MsgSomethingWrong)
		return
	}
	teacherID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("teacher_id")), 10, 64)
	txnID, err2 := strconv.ParseInt(strings.TrimSpace(r.FormValue("transaction_id")), 10, 64)
	if err != nil || err2 != nil || teacherID <= 0 || txnID <= 0 {
		sendErrorLog(w, ErrInvalidTeacherID.Error())
		return
	}
	affected, err := dbRW.GetQueries().RevokeTeacherPlanTransaction(ctx, queries.RevokeTeacherPlanTransactionParams{
		ID:        txnID,
		TeacherID: teacherID,
	})
	if err != nil || affected == 0 {
		sendErrorLog(w, "could not revoke plan grant")
		return
	}
	actor := auth.GetUser(ctx)
	insertAuditLogAs(ctx, actor, "plans", fmt.Sprintf(
		"revoked plan transaction #%d for teacher #%d",
		txnID, teacherID,
	))
	setSuccessFlash(w, "Plan grant revoked.")
	HttpRedirect(w, r, "/plans")
}

func buildPlansTeacherRows(ctx context.Context) ([]frontend.PlansTeacherRowItem, error) {
	rows, err := dbRO.GetQueries().GetAllTeachers(ctx)
	if err != nil {
		return nil, err
	}
	q := dbRO.GetQueries()
	items := make([]frontend.PlansTeacherRowItem, 0, len(rows))
	for _, row := range rows {
		name := utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName)
		tierLabel := "Free"
		plan, err := entitlements.TeacherPlan(ctx, q, row.ID)
		if err == nil && plan.IsPro {
			tierLabel = "Teacher Pro"
		}
		items = append(items, frontend.PlansTeacherRowItem{
			ID:       strconv.FormatInt(row.ID, 10),
			Name:     name,
			Email:    row.Email,
			PlanTier: tierLabel,
			Avatar:   buildTeacherListAvatarProps(row.ID, row.FirstName, row.MiddleName, row.LastName, row.AssignedColor, sql.NullString{}),
		})
	}
	return items, nil
}

func buildPlansTeacherViewModal(ctx context.Context, teacherID int64) (frontend.PlansTeacherViewModalData, error) {
	t, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		return frontend.PlansTeacherViewModalData{}, err
	}
	name := utils.ComposePersonName(t.FirstName, t.MiddleName, t.LastName)
	rows, err := dbRO.GetQueries().GetTeacherPlanTransactionsByTeacherID(ctx, queries.GetTeacherPlanTransactionsByTeacherIDParams{
		TeacherID: teacherID,
		Limit:     50,
		Offset:    0,
	})
	if err != nil {
		return frontend.PlansTeacherViewModalData{}, err
	}
	return frontend.PlansTeacherViewModalData{
		TeacherID:    strconv.FormatInt(teacherID, 10),
		TeacherName:  name,
		TeacherEmail: t.Email,
		Avatar:       planTeacherAvatar(ctx, teacherID),
		Summary:      buildPlanSummaryView(ctx, teacherID),
		Transactions: mapPlanTransactionItems(ctx, rows),
	}, nil
}

func buildPlansGrantModal(ctx context.Context, teacherID int64) (frontend.PlansGrantModalData, error) {
	t, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		return frontend.PlansGrantModalData{}, err
	}
	name := utils.ComposePersonName(t.FirstName, t.MiddleName, t.LastName)
	return frontend.PlansGrantModalData{
		TeacherID:   strconv.FormatInt(teacherID, 10),
		TeacherName: name,
		Avatar:      planTeacherAvatar(ctx, teacherID),
		Summary:     buildPlanSummaryView(ctx, teacherID),
	}, nil
}

func listAllRecentPlanTransactions(ctx context.Context) ([]frontend.PlanTransactionItem, error) {
	rows, err := dbRO.GetQueries().ListRecentTeacherPlanTransactions(ctx, queries.ListRecentTeacherPlanTransactionsParams{
		Limit:  plansRecentTransactionLimit,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	return mapPlanTransactionItems(ctx, rows), nil
}

func buildPlanSummaryView(ctx context.Context, teacherID int64) frontend.PlanSummaryView {
	plan, err := entitlements.TeacherPlan(ctx, dbRO.GetQueries(), teacherID)
	if err != nil {
		return frontend.PlanSummaryView{TierLabel: "Free"}
	}
	if !plan.IsPro {
		return frontend.PlanSummaryView{TierLabel: "Free", Detail: "No active Teacher Pro grant."}
	}
	label := "Teacher Pro"
	detail := ""
	switch plan.BillingKind {
	case constants.TeacherPlanBillingLifetime:
		detail = "Lifetime access"
	case constants.TeacherPlanBillingMonthly:
		detail = "Monthly grant"
	default:
		detail = string(plan.BillingKind)
	}
	if plan.EffectiveEnd != "" {
		detail += " · active until " + utils.FormatPlanDateTimePHT(plan.EffectiveEnd)
	}
	return frontend.PlanSummaryView{
		TierLabel:     label,
		Detail:        detail,
		ActiveGrantID: plan.GrantID,
	}
}

func mapPlanTransactionItems(ctx context.Context, rows []queries.TblTeacherPlanTransaction) []frontend.PlanTransactionItem {
	items := make([]frontend.PlanTransactionItem, 0, len(rows))
	for _, row := range rows {
		name := teacherNameByID(ctx, row.TeacherID)
		items = append(items, frontend.PlanTransactionItem{
			ID:             row.ID,
			TeacherID:      row.TeacherID,
			TeacherName:    name,
			TeacherAvatar:  planTeacherAvatar(ctx, row.TeacherID),
			BillingKind:    constants.TeacherPlanBillingKind(row.BillingKind),
			EffectiveStart: formatPlanTxnDate(row.EffectiveStart),
			EffectiveEnd:   formatPlanTxnEnd(utils.NullStringFromAny(row.EffectiveEnd)),
			Status:         planTransactionStatus(row),
			GrantedByName:  row.GrantedByName,
			Note:           row.Note,
			CreatedAt:      row.CreatedAt,
			CanRevoke:      planTransactionCanRevoke(row),
		})
	}
	return items
}

func planTransactionStatus(row queries.TblTeacherPlanTransaction) string {
	revoked := utils.NullStringFromAny(row.RevokedAt)
	if revoked.Valid && revoked.String != "" {
		return "Revoked"
	}
	now := time.Now().UTC()
	effectiveEnd := utils.NullStringFromAny(row.EffectiveEnd)
	if effectiveEnd.Valid && effectiveEnd.String != "" {
		end, err := time.Parse(constants.DateTimeSecondsLayout, effectiveEnd.String)
		if err == nil && !end.After(now) {
			return "Expired"
		}
	}
	start, err := time.Parse(constants.DateTimeSecondsLayout, row.EffectiveStart)
	if err == nil && start.After(now) {
		return "Scheduled"
	}
	if revoked.Valid {
		return "Revoked"
	}
	if effectiveEnd.Valid {
		end, err := time.Parse(constants.DateTimeSecondsLayout, effectiveEnd.String)
		if err == nil && !end.After(now) {
			return "Expired"
		}
	}
	return "Active"
}

func planTransactionCanRevoke(row queries.TblTeacherPlanTransaction) bool {
	revoked := utils.NullStringFromAny(row.RevokedAt)
	if revoked.Valid && revoked.String != "" {
		return false
	}
	return planTransactionStatus(row) == "Active" || planTransactionStatus(row) == "Scheduled"
}

func formatPlanTxnDate(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}

func formatPlanTxnEnd(end sql.NullString) string {
	if !end.Valid || end.String == "" {
		return "-"
	}
	return utils.FormatPlanDateTimePHT(end.String)
}

func planTeacherAvatar(ctx context.Context, teacherID int64) frontend.AvatarProps {
	t, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		name := teacherNameByID(ctx, teacherID)
		return buildTeacherListAvatarProps(teacherID, name, "", "", constants.DefaultTeacherAssignedColor, sql.NullString{})
	}
	return buildTeacherListAvatarProps(
		teacherID,
		t.FirstName,
		t.MiddleName,
		t.LastName,
		t.AssignedColor,
		t.ProfilePicture,
	)
}
