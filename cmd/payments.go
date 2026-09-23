package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/notifications"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func handlePaymentsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "payments", "/view"); ok {
		handlePaymentView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "payments", "/receipt"); ok {
		handlePaymentReceiptForm(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "payments", "/received"); ok {
		handlePaymentReceived(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "payments", "/defer"); ok {
		handlePaymentDefer(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handlePayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	role := auth.GetRole(r.Context())
	w.Header().Set("Content-Type", "text/html")
	if err := frontend.Payments(frontend.PaymentsData{
		ShowTeacherColumn: auth.HasAdminAccess(role),
	}).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePaymentsDatePresetPartial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	w.Header().Set("Content-Type", "text/html")
	if err := frontend.DatePresetForMonth(month, true).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePaymentsPartial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate, endDate, err := parseListDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	role := auth.GetRole(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if !auth.HasAdminAccess(role) {
		q = ""
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && !constants.ValidPaymentStatus(status) {
		HttpError(w, ErrInvalidStatus.Error(), http.StatusBadRequest)
		return
	}

	rows, err := loadPaymentRows(r.Context(), paymentsTeacherScope(role, auth.GetUser(r.Context()).ID), startDate, endDate, q, status)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	emptyMsg := "No payments found."
	if startDate == "" && endDate == "" && q == "" && status == "" {
		emptyMsg = "No payments recorded yet."
	}

	showTeacherColumn := auth.HasAdminAccess(role)

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.PaymentsPartial(rows, emptyMsg, showTeacherColumn).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func paymentsTeacherScope(role auth.Role, userID int64) int64 {
	if auth.IsTeacherScoped(role) {
		return userID
	}
	return 0
}

func loadPaymentRows(ctx context.Context, teacherID int64, startDate, endDate, q, status string) ([]frontend.PaymentRowData, error) {
	qNull := sql.NullString{String: q, Valid: q != ""}
	dbRows, err := dbRO.GetQueries().GetTeacherPaymentsFiltered(ctx, queries.GetTeacherPaymentsFilteredParams{
		Column1:     startDate,
		PeriodStart: startDate,
		Column3:     endDate,
		PeriodEnd:   endDate,
		Column5:     q,
		Column6:     qNull,
		Column7:     qNull,
		Column8:     qNull,
		Column9:     status,
		Status:      status,
		Column11:    teacherID,
		TeacherID:   teacherID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load payments")
	}

	teacherIDs := make([]int64, 0, len(dbRows))
	for _, row := range dbRows {
		teacherIDs = append(teacherIDs, row.TeacherID)
	}
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to load teacher roles")
	}

	rows := make([]frontend.PaymentRowData, 0, len(dbRows))
	for _, row := range dbRows {
		rows = append(rows, mapPaymentRow(row, rolesMap, teacherID))
	}
	return rows, nil
}

func mapPaymentRow(row queries.GetTeacherPaymentsFilteredRow, rolesMap map[int64][]constants.TeacherRole, scopedTeacherID int64) frontend.PaymentRowData {
	teacherName := utils.ComposePersonName(row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName)
	receivedAt := ""
	if row.ReceivedAt != nil {
		receivedAt = formatPaymentTimestamp(fmt.Sprint(row.ReceivedAt))
	}
	status := constants.PaymentStatus(row.Status)
	return frontend.PaymentRowData{
		ID:              strconv.FormatInt(row.ID, 10),
		TeacherName:     teacherName,
		TeacherAvatar: avatarWithTeacherRoles(
			buildTeacherListAvatarProps(row.TeacherID, row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName, constants.DefaultTeacherAssignedColor, row.TeacherProfilePicture),
			rolesMap[row.TeacherID],
		),
		PeriodLabel:        formatReportCutoffLabel(row.PeriodStart, row.PeriodEnd),
		PaymentMethod:      constants.PaymentMethod(row.PaymentMethod),
		ReferenceNumber:    row.ReferenceNumber,
		Amount:             row.Amount,
		Currency:           row.Currency,
		Status:             status,
		SentByName:         row.SentByName,
		SentAt:             formatPaymentTimestamp(row.SentAt),
		ReceivedAt:         receivedAt,
		CanConfirmReceived: scopedTeacherID > 0 && status == constants.PaymentStatusPending,
	}
}

func loadPaymentRow(ctx context.Context, paymentID int64, scopedTeacherID int64) (frontend.PaymentRowData, error) {
	payment, err := dbRO.GetQueries().GetTeacherPaymentByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return frontend.PaymentRowData{}, ErrPaymentNotFound
		}
		return frontend.PaymentRowData{}, fmt.Errorf("failed to load payment")
	}
	if scopedTeacherID > 0 && payment.TeacherID != scopedTeacherID {
		return frontend.PaymentRowData{}, ErrForbidden
	}

	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, payment.TeacherID)
	if err != nil {
		return frontend.PaymentRowData{}, fmt.Errorf("failed to load teacher")
	}

	rolesMap, err := loadRolesByTeacherIDs(ctx, []int64{payment.TeacherID})
	if err != nil {
		return frontend.PaymentRowData{}, fmt.Errorf("failed to load teacher roles")
	}

	filteredRow := queries.GetTeacherPaymentsFilteredRow{
		ID:                  payment.ID,
		TeacherID:           payment.TeacherID,
		SentByName:          payment.SentByName,
		PaymentMethod:       payment.PaymentMethod,
		ReferenceNumber:     payment.ReferenceNumber,
		Amount:              payment.Amount,
		Currency:            payment.Currency,
		PeriodStart:         payment.PeriodStart,
		PeriodEnd:           payment.PeriodEnd,
		Status:              payment.Status,
		SentAt:              payment.SentAt,
		ReceivedAt:          payment.ReceivedAt,
		TeacherFirstName:    profile.FirstName,
		TeacherMiddleName:   profile.MiddleName,
		TeacherLastName:     profile.LastName,
		TeacherProfilePicture: profile.ProfilePicture,
	}
	return mapPaymentRow(filteredRow, rolesMap, scopedTeacherID), nil
}

func teacherHasPaymentForPeriod(ctx context.Context, teacherID int64, startDate, endDate string) (bool, error) {
	hasPayment, err := dbRO.GetQueries().TeacherHasPaymentForPeriod(ctx, queries.TeacherHasPaymentForPeriodParams{
		TeacherID:   teacherID,
		PeriodStart: startDate,
		PeriodEnd:   endDate,
	})
	if err != nil {
		return false, err
	}
	return hasPayment > 0, nil
}

func paymentSentDisabledTooltip(startDate, endDate string) string {
	nextStart, nextEnd, ok := utils.NextCutoffDates(startDate, endDate)
	if !ok {
		return "Payment already sent for this cutoff period."
	}
	return fmt.Sprintf(
		"Payment already sent for this cutoff. Available again on the next cutoff (%s).",
		formatReportCutoffLabel(nextStart, nextEnd),
	)
}

func loadPaymentStatusesForPeriod(ctx context.Context, startDate, endDate string) (map[int64]constants.PaymentStatus, error) {
	if startDate == "" || endDate == "" {
		return map[int64]constants.PaymentStatus{}, nil
	}
	statusRows, err := dbRO.GetQueries().GetTeacherPaymentStatusesForPeriod(ctx, queries.GetTeacherPaymentStatusesForPeriodParams{
		PeriodStart: startDate,
		PeriodEnd:   endDate,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[int64]constants.PaymentStatus, len(statusRows))
	for _, row := range statusRows {
		if _, ok := out[row.TeacherID]; ok {
			continue
		}
		out[row.TeacherID] = constants.PaymentStatus(row.Status)
	}
	return out, nil
}

func handleReportPaymentForm(w http.ResponseWriter, r *http.Request, teacherID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if hasPayment, err := teacherHasPaymentForPeriod(ctx, teacherID, startDate, endDate); err != nil {
		sendErrorLog(w, "failed to check payment status")
		return
	} else if hasPayment {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showErrorBanner":%q}`, paymentSentDisabledTooltip(startDate, endDate)))
		return
	}

	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		HttpError(w, "Teacher not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	teacherName := utils.ComposePersonName(profile.FirstName, profile.MiddleName, profile.LastName)
	frontend.ReportPaymentModal(frontend.ReportPaymentFormData{
		TeacherID:     strconv.FormatInt(teacherID, 10),
		TeacherName:   teacherName,
		TeacherAvatar: buildReportTeacherAvatarProps(teacherID, profile),
		StartDate:     startDate,
		EndDate:       endDate,
		CutoffLabel:   formatReportCutoffLabel(startDate, endDate),
		TotalEarnings: loadTeacherReportEarnings(ctx, teacherID, startDate, endDate),
	}).Render(ctx, w)
}

func handleReportPaymentSubmit(w http.ResponseWriter, r *http.Request, teacherID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, "invalid request")
		return
	}

	startDate := strings.TrimSpace(r.FormValue("startDate"))
	endDate := strings.TrimSpace(r.FormValue("endDate"))
	paymentMethod := strings.TrimSpace(r.FormValue("paymentMethod"))
	referenceNumber := strings.TrimSpace(r.FormValue("referenceNumber"))
	currency := strings.TrimSpace(r.FormValue("currency"))
	amountRaw := strings.TrimSpace(r.FormValue("amount"))

	if startDate == "" || endDate == "" {
		renderReportPaymentError(w, r, teacherID, startDate, endDate, ErrMissingDateRange.Error())
		return
	}
	if !constants.ValidPaymentMethod(paymentMethod) {
		renderReportPaymentError(w, r, teacherID, startDate, endDate, ErrInvalidPaymentMethod.Error())
		return
	}
	if referenceNumber == "" {
		renderReportPaymentError(w, r, teacherID, startDate, endDate, ErrReferenceNumberRequired.Error())
		return
	}
	if !constants.ValidCurrency(currency) {
		renderReportPaymentError(w, r, teacherID, startDate, endDate, ErrInvalidCurrency.Error())
		return
	}
	amount, err := strconv.ParseFloat(amountRaw, 64)
	if err != nil || amount <= 0 {
		renderReportPaymentError(w, r, teacherID, startDate, endDate, ErrInvalidPaymentAmount.Error())
		return
	}

	ctx := r.Context()
	if hasPayment, err := teacherHasPaymentForPeriod(ctx, teacherID, startDate, endDate); err != nil {
		sendErrorLog(w, "failed to check payment status")
		return
	} else if hasPayment {
		renderReportPaymentError(w, r, teacherID, startDate, endDate, paymentSentDisabledTooltip(startDate, endDate))
		return
	}

	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		sendErrorLog(w, "teacher not found")
		return
	}

	user := auth.GetUser(ctx)
	var sentByID interface{}
	if user.ID > 0 {
		sentByID = user.ID
	}
	if err := dbRW.GetQueries().InsertTeacherPayment(ctx, queries.InsertTeacherPaymentParams{
		TeacherID:       teacherID,
		SentByTeacherID: sentByID,
		SentByName:      user.Name,
		PaymentMethod:   paymentMethod,
		ReferenceNumber: referenceNumber,
		Amount:          amount,
		Currency:        currency,
		PeriodStart:     startDate,
		PeriodEnd:       endDate,
	}); err != nil {
		sendErrorLog(w, "failed to send payment")
		return
	}

	teacherName := utils.ComposePersonName(profile.FirstName, profile.MiddleName, profile.LastName)
	insertAuditLogAs(ctx, user, "payments", fmt.Sprintf(
		"sent payment to %s (%s %s, ref %s)",
		teacherName,
		utils.FormatCurrency(amount, currency),
		currency,
		referenceNumber,
	))
	notifyTeacher(ctx, teacherID, teacherName, user, notifications.KindPaymentSent,
		fmt.Sprintf("Payment of %s was sent to you (ref %s)", utils.FormatCurrency(amount, currency), referenceNumber), "")

	row, err := loadReportRow(ctx, teacherID, startDate, endDate)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"showSuccessBanner":"Payment sent"}`)
		windowCloseReportPaymentModal(w)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("HX-Trigger", `{"showSuccessBanner":"Payment sent"}`)
	if err := frontend.ReportPaymentPostResponse(row, startDate, endDate).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderReportPaymentError(w http.ResponseWriter, r *http.Request, teacherID int64, startDate, endDate, message string) {
	ctx := r.Context()
	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		sendErrorLog(w, "teacher not found")
		return
	}
	w.Header().Set("Content-Type", "text/html")
	teacherName := utils.ComposePersonName(profile.FirstName, profile.MiddleName, profile.LastName)
	frontend.ReportPaymentModal(frontend.ReportPaymentFormData{
		TeacherID:     strconv.FormatInt(teacherID, 10),
		TeacherName:   teacherName,
		TeacherAvatar: buildReportTeacherAvatarProps(teacherID, profile),
		StartDate:     startDate,
		EndDate:       endDate,
		CutoffLabel:   formatReportCutoffLabel(startDate, endDate),
		TotalEarnings: loadTeacherReportEarnings(ctx, teacherID, startDate, endDate),
		Error:         message,
	}).Render(ctx, w)
}

func loadTeacherReportEarnings(ctx context.Context, teacherID int64, startDate, endDate string) []frontend.CurrencyTotal {
	rows, err := dbRO.GetQueries().GetReportTeacherEarnings(ctx, reportEarningsParams("", startDate, endDate, reportRoleFilters{}))
	if err != nil {
		return nil
	}
	earnings := make([]frontend.CurrencyTotal, 0)
	for _, row := range rows {
		if row.TeacherID != teacherID {
			continue
		}
		total := sqlNumericToFloat64(row.TotalRate)
		if total == 0 {
			continue
		}
		earnings = append(earnings, frontend.CurrencyTotal{
			Currency: row.Currency,
			Total:    total,
		})
	}
	return earnings
}

func windowCloseReportPaymentModal(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div id="reportPaymentModalHost" hx-swap-oob="innerHTML"></div>`)
}

func handlePaymentView(w http.ResponseWriter, r *http.Request, paymentID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)
	row, err := loadPaymentRow(ctx, paymentID, paymentsTeacherScope(role, user.ID))
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			sendErrorLog(w, ErrPaymentNotFound.Error())
			return
		}
		if errors.Is(err, ErrForbidden) {
			HttpError(w, "Forbidden", http.StatusForbidden)
			return
		}
		sendErrorLog(w, "failed to load payment")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.PaymentViewModal(frontend.PaymentViewData{
		ShowTeacher:     auth.HasAdminAccess(role),
		TeacherName:     row.TeacherName,
		TeacherAvatar:   row.TeacherAvatar,
		PeriodLabel:     row.PeriodLabel,
		PaymentMethod:   row.PaymentMethod,
		ReferenceNumber: row.ReferenceNumber,
		Amount:          row.Amount,
		Currency:        row.Currency,
		Status:          row.Status,
		SentByName:      row.SentByName,
		SentAt:          row.SentAt,
		ReceivedAt:      row.ReceivedAt,
	}).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePaymentReceiptForm(w http.ResponseWriter, r *http.Request, paymentID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if !auth.IsTeacherScoped(user.Role) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	payment, err := dbRO.GetQueries().GetTeacherPaymentByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorLog(w, ErrPaymentNotFound.Error())
			return
		}
		sendErrorLog(w, "failed to load payment")
		return
	}
	if payment.TeacherID != user.ID {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}
	if payment.Status != string(constants.PaymentStatusPending) {
		sendErrorLog(w, "payment is not pending")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.PaymentReceiptModal(frontend.PaymentReceiptData{
		PaymentID:       strconv.FormatInt(payment.ID, 10),
		PaymentMethod:   constants.PaymentMethod(payment.PaymentMethod),
		ReferenceNumber: payment.ReferenceNumber,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		PeriodLabel:     formatReportCutoffLabel(payment.PeriodStart, payment.PeriodEnd),
		SentAt:          formatPaymentTimestamp(payment.SentAt),
		ShowDeferAction: false,
	}).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePaymentReceived(w http.ResponseWriter, r *http.Request, paymentID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(user.Role) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	payment, err := dbRO.GetQueries().GetTeacherPaymentByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorLog(w, ErrPaymentNotFound.Error())
			return
		}
		sendErrorLog(w, "failed to load payment")
		return
	}
	if payment.TeacherID != user.ID {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}
	if payment.Status != string(constants.PaymentStatusPending) {
		if r.Header.Get("HX-Request") != "" {
			sendErrorLog(w, "payment is not pending")
			return
		}
		HttpRedirect(w, r, "/dashboard")
		return
	}

	if err := dbRW.GetQueries().MarkTeacherPaymentReceived(ctx, queries.MarkTeacherPaymentReceivedParams{
		ID:        paymentID,
		TeacherID: user.ID,
	}); err != nil {
		sendErrorLog(w, "failed to confirm payment")
		return
	}

	if r.Header.Get("HX-Request") != "" && r.Header.Get("X-Payment-Source") == "history" {
		row, err := loadPaymentRow(ctx, paymentID, user.ID)
		if err != nil {
			sendErrorLog(w, "failed to load payment row")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("HX-Trigger", `{"showSuccessBanner":"Payment marked as received"}`)
		if err := frontend.PaymentReceivedPostResponse(row, false).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", utils.URL("/dashboard"))
		return
	}

	HttpRedirect(w, r, "/dashboard")
}

func handlePaymentDefer(w http.ResponseWriter, r *http.Request, paymentID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(user.Role) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	payment, err := dbRO.GetQueries().GetTeacherPaymentByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorLog(w, ErrPaymentNotFound.Error())
			return
		}
		sendErrorLog(w, "failed to load payment")
		return
	}
	if payment.TeacherID != user.ID {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}
	if payment.Status != string(constants.PaymentStatusPending) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	accessID, err := dbRO.GetQueries().GetLatestOpenAccessByTeacherID(ctx, user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logs.Log().Warn("payment defer access lookup", zap.Error(err))
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := dbRW.GetQueries().DeferTeacherPaymentReceipt(ctx, queries.DeferTeacherPaymentReceiptParams{
		DismissedAccessID: accessID,
		ID:                paymentID,
		TeacherID:         user.ID,
	}); err != nil {
		sendErrorLog(w, "failed to defer payment receipt")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func populatePaymentReceipt(ctx context.Context, data *frontend.DashboardData, teacherID int64) {
	accessID, err := dbRO.GetQueries().GetLatestOpenAccessByTeacherID(ctx, teacherID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logs.Log().Warn("payment receipt access lookup", zap.Error(err))
		return
	}

	payment, err := dbRO.GetQueries().GetPendingPaymentForTeacherReceipt(ctx, queries.GetPendingPaymentForTeacherReceiptParams{
		TeacherID:         teacherID,
		Column2:           accessID,
		DismissedAccessID: accessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		logs.Log().Warn("payment receipt lookup", zap.Error(err))
		return
	}

	data.ShowPaymentReceipt = true
	data.PaymentReceipt = frontend.PaymentReceiptData{
		PaymentID:       strconv.FormatInt(payment.ID, 10),
		PaymentMethod:   constants.PaymentMethod(payment.PaymentMethod),
		ReferenceNumber: payment.ReferenceNumber,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		PeriodLabel:     formatReportCutoffLabel(payment.PeriodStart, payment.PeriodEnd),
		SentAt:          formatPaymentTimestamp(payment.SentAt),
		ShowDeferAction: true,
	}
}

func formatPaymentTimestamp(raw string) string {
	if raw == "" {
		return ""
	}
	t, err := time.ParseInLocation(constants.DateTimeSecondsLayout, raw, time.UTC)
	if err != nil {
		t, err = time.ParseInLocation(constants.DateTimeLayout, raw, time.UTC)
		if err != nil {
			return raw
		}
	}
	return utils.DateTimeSecondsPHT(t)
}
