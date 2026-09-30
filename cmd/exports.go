package cmd

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/entitlements"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func handleExports(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	plan, err := entitlements.TeacherPlan(ctx, dbRO.GetQueries(), user.ID)
	if err != nil {
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	writeHTML(w)
	if err := frontend.Exports(frontend.ExportsPageData{IsPro: plan.IsPro}).Render(ctx, w); err != nil {
		logs.Log().Error("render exports", zap.Error(err))
	}
}

func requireTeacherProExport(w http.ResponseWriter, r *http.Request) (auth.User, bool) {
	if auth.GetRole(r.Context()) != auth.RoleTeacher {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return auth.User{}, false
	}
	user := auth.GetUser(r.Context())
	ok, err := entitlements.HasFeature(r.Context(), dbRO.GetQueries(), user, auth.RoleTeacher, entitlements.FeatureExportClasses)
	if err != nil {
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return auth.User{}, false
	}
	if !ok {
		HttpError(w, MsgEntitlementsUpgradeRequired, http.StatusForbidden)
		return auth.User{}, false
	}
	return user, true
}

func handleExportClasses(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	user, ok := requireTeacherProExport(w, r)
	if !ok {
		return
	}
	startDate := utils.NormalizeDatePHT(r.URL.Query().Get("startDate"))
	endDate := utils.NormalizeDatePHT(r.URL.Query().Get("endDate"))
	if startDate == "" || endDate == "" {
		HttpError(w, ErrMissingDateRange.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	overdueCutoff := overdueCutoffPHT(ctx)
	rows, err := dbRO.GetQueries().GetClassesListFiltered(ctx, classesListListParams(user.ID, startDate, endDate, "", "", overdueCutoff, 10000, 0))
	if err != nil {
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	filename := fmt.Sprintf("classes_%s_%s.csv", startDate, endDate)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"date", "start_time", "end_time", "student", "status", "rate", "currency"})
	for _, row := range rows {
		_ = cw.Write([]string{
			row.Date,
			row.StartTime.String,
			row.EndTime.String,
			row.StudentName,
			row.Status,
			fmt.Sprintf("%.2f", row.Rate),
			row.Currency,
		})
	}
	cw.Flush()
	insertAuditLogAs(ctx, user, "exports", fmt.Sprintf("exported classes CSV (%s to %s)", startDate, endDate))
}

func handleExportStudents(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	user, ok := requireTeacherProExport(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	rows, err := dbRO.GetQueries().GetStudentsByTeacherID(ctx, user.ID)
	if err != nil {
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=my_students.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"name", "status", "contact", "currency", "rate"})
	for _, row := range rows {
		contact := ""
		if row.Contact.Valid {
			contact = row.Contact.String
		}
		_ = cw.Write([]string{
			row.Name,
			row.Status,
			contact,
			row.Currency,
			fmt.Sprintf("%.2f", row.RatePerClass),
		})
	}
	cw.Flush()
	insertAuditLogAs(ctx, user, "exports", "exported students CSV")
}

func handleExportSchedule(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	user, ok := requireTeacherProExport(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	start := utils.TodayPHT()
	endT, err := utils.ParseDatePHT(start)
	if err != nil || endT == nil {
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	end := utils.DatePHT(endT.AddDate(0, 0, 90))
	overdueCutoff := overdueCutoffPHT(ctx)
	rows, err := dbRO.GetQueries().GetClassesListFiltered(ctx, classesListListParams(
		user.ID, start, end, string(constants.ClassListFilterScheduled), "", overdueCutoff, 10000, 0,
	))
	if err != nil {
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=schedule.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"date", "start_time", "end_time", "student", "duration_minutes", "status"})
	for _, row := range rows {
		if strings.EqualFold(row.Status, string(constants.ClassStatusConducted)) {
			continue
		}
		_ = cw.Write([]string{
			row.Date,
			row.StartTime.String,
			row.EndTime.String,
			row.StudentName,
			fmt.Sprintf("%d", row.DurationMinutes),
			row.Status,
		})
	}
	cw.Flush()
	insertAuditLogAs(ctx, user, "exports", "exported schedule CSV")
}
