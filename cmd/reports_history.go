package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

func handleReportsHistory(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeHTML(w)
	if err := frontend.ReportsHistory().Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportsHistoryDatePresetPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	writeHTML(w)
	if err := frontend.DatePresetForMonth(month, true).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportsHistoryPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	startDate, endDate, err := parseListDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := loadReportHistoryRows(r.Context(), startDate, endDate, q)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	emptyMsg := "No report generations found."
	if startDate == "" && endDate == "" && q == "" {
		emptyMsg = "No reports generated yet."
	}

	writeHTML(w)
	if err := frontend.ReportsHistoryPartial(rows, emptyMsg).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadReportHistoryRows(ctx context.Context, startDate, endDate, q string) ([]frontend.ReportHistoryRowData, error) {
	qNull := sql.NullString{String: q, Valid: q != ""}
	dbRows, err := dbRO.GetQueries().GetReportGenerationsFiltered(ctx, queries.GetReportGenerationsFilteredParams{
		Column1:   startDate,
		StartDate: startDate,
		Column3:   endDate,
		EndDate:   endDate,
		Column5:   q,
		Column6:   qNull,
		Column7:   qNull,
		Column8:   qNull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load report history")
	}

	teacherIDs := make([]int64, 0, len(dbRows))
	for _, row := range dbRows {
		teacherIDs = append(teacherIDs, row.TeacherID)
	}
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to load teacher roles")
	}

	rows := make([]frontend.ReportHistoryRowData, 0, len(dbRows))
	for _, row := range dbRows {
		rows = append(rows, mapReportHistoryRow(ctx, row, rolesMap))
	}
	return rows, nil
}

func mapReportHistoryRow(ctx context.Context, row queries.GetReportGenerationsFilteredRow, rolesMap map[int64][]constants.TeacherRole) frontend.ReportHistoryRowData {
	teacherName := utils.ComposePersonName(row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName)
	item := frontend.ReportHistoryRowData{
		TeacherName: teacherName,
		TeacherAvatar: avatarWithTeacherRoles(
			buildTeacherListAvatarProps(row.TeacherID, row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName, constants.DefaultTeacherAssignedColor, row.TeacherProfilePicture),
			rolesMap[row.TeacherID],
		),
		PeriodLabel: formatReportCutoffLabel(row.StartDate, row.EndDate),
		RecordCount: row.RecordCount,
		GeneratedAt: formatPaymentTimestamp(row.GeneratedAt),
	}
	if filename, ok := reportCacheAvailable(ctx, row.OutputPath); ok {
		item.DownloadReady = true
		item.Filename = filename
	}
	return item
}
