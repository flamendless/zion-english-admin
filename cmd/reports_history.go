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
	kindFilter := parseReportHistoryKindFilter(r)
	page := utils.ParsePageQuery(r)
	allRows, err := loadReportHistoryRows(r.Context(), startDate, endDate, q, kindFilter)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page.Total = int64(len(allRows))
	rows := paginateSlice(allRows, page)

	emptyMsg := reportHistoryEmptyMessage(startDate, endDate, q, kindFilter)

	pagination := frontend.BuildPaginationData(page.Number, page.Size, page.Total)
	partialsURL := utils.URL("/reports/history/partials/rows")
	includeSelector := "#reportsHistoryFilters"
	targetSelector := "#reportsHistoryTableBody"

	writeHTML(w)
	if err := frontend.ReportsHistoryPartial(rows, emptyMsg, pagination, partialsURL, includeSelector, targetSelector).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseReportHistoryKindFilter(r *http.Request) string {
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	if kind == string(constants.ReportGenerationKindSummary) {
		return kind
	}
	return ""
}

func reportHistoryEmptyMessage(startDate, endDate, q, kindFilter string) string {
	if kindFilter == string(constants.ReportGenerationKindSummary) {
		return "No payroll summary reports found."
	}
	if startDate == "" && endDate == "" && q == "" && kindFilter == "" {
		return "No reports generated yet."
	}
	return "No report generations found."
}

func loadReportHistoryRows(ctx context.Context, startDate, endDate, q, kindFilter string) ([]frontend.ReportHistoryRowData, error) {
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
		Column9:   qNull,
		Column10:  kindFilter,
		Kind:      kindFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load report history")
	}

	teacherIDs := make([]int64, 0, len(dbRows))
	for _, row := range dbRows {
		if row.Kind == string(constants.ReportGenerationKindSummary) || !row.TeacherID.Valid {
			continue
		}
		teacherIDs = append(teacherIDs, row.TeacherID.Int64)
	}
	uniqueIDs := uniqueTeacherIDs(teacherIDs)
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load teacher roles")
	}
	statusMap, err := loadTeacherStatusesByIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load teacher statuses")
	}

	rows := make([]frontend.ReportHistoryRowData, 0, len(dbRows))
	for _, row := range dbRows {
		rows = append(rows, mapReportHistoryRow(ctx, row, rolesMap, statusMap))
	}
	return rows, nil
}

func mapReportHistoryRow(ctx context.Context, row queries.GetReportGenerationsFilteredRow, rolesMap map[int64][]constants.TeacherRole, statusMap map[int64]constants.TeacherStatus) frontend.ReportHistoryRowData {
	item := frontend.ReportHistoryRowData{
		Kind:        constants.ReportGenerationKind(row.Kind),
		PeriodLabel: formatReportCutoffLabel(row.StartDate, row.EndDate),
		RecordCount: row.RecordCount,
		GeneratedAt: formatPaymentTimestamp(row.GeneratedAt),
	}
	if row.Kind == string(constants.ReportGenerationKindSummary) {
		item.TeacherName = constants.ReportHistorySummaryLabel
	} else {
		teacherID := row.TeacherID.Int64
		teacherName := utils.ComposePersonName(
			nullStringValue(row.TeacherFirstName),
			nullStringValue(row.TeacherMiddleName),
			nullStringValue(row.TeacherLastName),
		)
		item.TeacherName = teacherName
		item.TeacherAvatar = avatarWithTeacherRoles(
			buildTeacherListAvatarProps(teacherID, nullStringValue(row.TeacherFirstName), nullStringValue(row.TeacherMiddleName), nullStringValue(row.TeacherLastName), constants.DefaultTeacherAssignedColor, row.TeacherProfilePicture),
			rolesMap[teacherID],
			teacherStatusFromMap(statusMap, teacherID),
		)
	}
	if filename, ok := reportCacheAvailable(ctx, row.OutputPath); ok {
		item.DownloadReady = true
		item.Filename = filename
	}
	return item
}
