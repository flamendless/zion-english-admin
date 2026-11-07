package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/billing"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/processor"
	"zion-english/internal/reports"
	"zion-english/internal/storage"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

type reportEarningJSON struct {
	Currency string  `json:"currency"`
	Total    float64 `json:"total"`
}

func handleReportsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "reports", "/payment"); ok {
		if r.Method == http.MethodGet {
			handleReportPaymentForm(w, r, id)
			return
		}
		if r.Method == http.MethodPost {
			handleReportPaymentSubmit(w, r, id)
			return
		}
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}
	if id, ok := extractPathID(r, "reports", "/view"); ok {
		handleReportView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "reports", "/generate"); ok {
		handleReportGenerate(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleReports(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeHTML(w)
	if err := frontend.Reports().Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportsDatePresetPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	month := strings.TrimSpace(r.URL.Query().Get("month"))
	writeHTML(w)
	if err := frontend.DatePresetForMonth(month, true).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportsSummaryDatePresetPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	month := strings.TrimSpace(r.URL.Query().Get("month"))
	writeHTML(w)
	if err := frontend.DatePresetForMonthScoped(month, true, "summaryDatePresetGroup", "summaryDatePreset").Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportsPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	roleFilters := parseReportRoleFilters(r)
	sort := parseListSort(r, frontend.ListSortKindReport)
	page := utils.ParsePageQuery(r)
	allRows, err := loadReportRows(r.Context(), startDate, endDate, q, roleFilters)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sortReportRows(allRows, sort)
	page.Total = int64(len(allRows))
	rows := paginateSlice(allRows, page)

	emptyMsg := "No teachers found."
	hasDateRange := startDate != "" && endDate != ""
	if !hasDateRange {
		emptyMsg = "Select a date range."
	}

	summary := frontend.ReportsSummaryData{
		HasDateRange: hasDateRange,
		CutoffLabel:  formatReportCutoffLabel(startDate, endDate),
		Earnings:     aggregateReportEarnings(allRows),
	}

	pagination := frontend.BuildPaginationData(page.Number, page.Size, page.Total)
	partialsURL := utils.URL("/reports/partials/rows")
	includeSelector := "#reportsFilters"
	targetSelector := "#reportsTableBody"

	renderHTMLComponent(r.Context(), w, "render reports partial", frontend.ReportsPartial(rows, startDate, endDate, emptyMsg, summary, pagination, partialsURL, includeSelector, targetSelector))
}

func handleReportsAllTeachers(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	roleFilters := parseReportRoleFilters(r)
	rows, err := loadReportRows(r.Context(), startDate, endDate, "", roleFilters)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sortReportRows(rows, parseListSort(r, frontend.ListSortKindReport))

	modalRows := make([]frontend.ReportAllTeachersRow, 0, len(rows))
	for _, row := range rows {
		modalRows = append(modalRows, frontend.ReportAllTeachersRow{
			TeacherName:      row.TeacherName,
			TeacherAvatar:    row.TeacherAvatar,
			ConductedClasses: row.ConductedClasses,
			CancelledClasses: row.CancelledClasses,
			TotalClasses:     row.TotalClasses,
			Earnings:         row.Earnings,
		})
	}

	emptyMsg := "No teachers found."
	if startDate == "" || endDate == "" {
		emptyMsg = "Select a date range."
	}

	writeHTML(w)
	if err := frontend.ReportAllTeachersModal(frontend.ReportAllTeachersData{
		CutoffLabel: formatReportCutoffLabel(startDate, endDate),
		Rows:        modalRows,
		EmptyMsg:    emptyMsg,
	}).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportSummary(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := constants.ParseSummaryReportKind(strings.TrimSpace(r.URL.Query().Get("kind")))
	if kind == constants.SummaryReportKindTeachersPayroll {
		handleTeachersPayrollSummary(w, r, startDate, endDate)
		return
	}
	handleParentsSummaryReport(w, r, startDate, endDate)
}

func handleReportSummaryView(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	kind := constants.ParseSummaryReportKind(strings.TrimSpace(r.URL.Query().Get("kind")))
	if kind == constants.SummaryReportKindTeachersPayroll {
		handleTeachersPayrollSummaryView(w, r, startDate, endDate)
		return
	}
	handleParentsSummaryView(w, r, startDate, endDate)
}

func loadParentsSummarySheets(ctx context.Context, q, startDate, endDate string, roleFilters reportRoleFilters) ([]queries.GetReportTeacherSummariesRow, []processor.SummaryTeacherSheet, []queries.GetReportSummaryRowsRow, error) {
	summaries, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, reportSearchParams(q, startDate, endDate, roleFilters))
	if err != nil {
		return nil, nil, nil, err
	}
	if len(summaries) == 0 {
		return nil, nil, nil, errNoTeachersForSummary
	}

	summaryRows, err := dbRO.GetQueries().GetReportSummaryRows(ctx, reportSummaryParams(q, startDate, endDate, roleFilters))
	if err != nil {
		return nil, nil, nil, err
	}

	sheets := buildSummaryTeacherSheets(summaries, summaryRows)
	return summaries, sheets, summaryRows, nil
}

func handleParentsSummaryReport(w http.ResponseWriter, r *http.Request, startDate, endDate string) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	roleFilters := parseReportRoleFilters(r)
	ctx := r.Context()

	_, sheets, summaryRows, err := loadParentsSummarySheets(ctx, q, startDate, endDate, roleFilters)
	if err != nil {
		if errors.Is(err, errNoTeachersForSummary) {
			HttpError(w, "No teachers found.", http.StatusBadRequest)
			return
		}
		logs.Log().Error("load report summaries for summary export", zap.Error(err))
		HttpError(w, "Failed to load report data", http.StatusInternalServerError)
		return
	}
	filename := fmt.Sprintf("summary_%s_%s_%s.xlsx", startDate, endDate, utils.RandomString(8))
	outputPath := filepath.Join("tmp", filename)
	if err := processor.SaveSummaryReport(sheets, outputPath); err != nil {
		logs.Log().Error("save summary report xlsx", zap.Error(err))
		HttpError(w, "Failed to generate summary report", http.StatusInternalServerError)
		return
	}

	if err := putReportFile(ctx, filename, outputPath); err != nil {
		logs.Log().Error("upload summary report to storage", zap.Error(err))
		HttpError(w, "Failed to save summary report", http.StatusInternalServerError)
		return
	}

	existing, existingErr := dbRO.GetQueries().GetSummaryReportGeneration(ctx, queries.GetSummaryReportGenerationParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if existingErr == nil && existing.OutputPath != "" {
		oldBase := reportOutputBasename(existing.OutputPath)
		if oldBase != "" && oldBase != filename {
			_ = storage.Default().Delete(ctx, storage.CategoryReports, oldBase)
			_ = os.Remove(filepath.Join("tmp", oldBase))
		}
	}

	if err := dbRW.GetQueries().UpsertSummaryReportGeneration(ctx, queries.UpsertSummaryReportGenerationParams{
		StartDate:   startDate,
		EndDate:     endDate,
		ContentHash: string(constants.ReportGenerationKindSummary),
		OutputPath:  filename,
		RecordCount: int64(len(summaryRows)),
	}); err != nil {
		logs.Log().Error("upsert summary report generation", zap.Error(err))
		HttpError(w, "Failed to save summary report", http.StatusInternalServerError)
		return
	}

	user := auth.GetUser(ctx)
	insertAuditLogAs(ctx, user, "reports", fmt.Sprintf(
		"generated parents summary report (%s to %s): %s",
		startDate, endDate, filename,
	))

	if !serveReportDownload(w, r, filename) {
		HttpError(w, "Failed to download summary report", http.StatusInternalServerError)
	}
}

func handleParentsSummaryView(w http.ResponseWriter, r *http.Request, startDate, endDate string) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	roleFilters := parseReportRoleFilters(r)
	ctx := r.Context()

	summaries, sheets, _, err := loadParentsSummarySheets(ctx, q, startDate, endDate, roleFilters)
	if err != nil {
		if errors.Is(err, errNoTeachersForSummary) {
			HttpError(w, "No teachers found.", http.StatusBadRequest)
			return
		}
		logs.Log().Error("load parents summary for view", zap.Error(err))
		HttpError(w, "Failed to load report data", http.StatusInternalServerError)
		return
	}

	writeHTML(w)
	if err := frontend.ReportSummaryViewModal(buildParentsSummaryViewData(summaries, sheets, startDate, endDate)).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTeachersPayrollSummaryView(w http.ResponseWriter, r *http.Request, startDate, endDate string) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	roleFilters := parseReportRoleFilters(r)
	ctx := r.Context()

	payrollRows, err := buildTeachersPayrollRows(ctx, q, startDate, endDate, roleFilters)
	if err != nil {
		logs.Log().Error("build teachers payroll rows for view", zap.Error(err))
		HttpError(w, "Failed to load report data", http.StatusInternalServerError)
		return
	}
	if len(payrollRows) == 0 {
		HttpError(w, "No class data for the selected range.", http.StatusBadRequest)
		return
	}

	summaries, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, reportSearchParams(q, startDate, endDate, roleFilters))
	if err != nil {
		logs.Log().Error("load teacher summaries for payroll view", zap.Error(err))
		HttpError(w, "Failed to load report data", http.StatusInternalServerError)
		return
	}
	avatarByTeacher := map[int64]frontend.AvatarProps{}
	for _, summary := range summaries {
		avatarByTeacher[summary.TeacherID] = buildReportSummaryAvatarProps(summary)
	}

	writeHTML(w)
	if err := frontend.ReportSummaryViewModal(buildTeachersPayrollSummaryViewData(payrollRows, avatarByTeacher, startDate, endDate)).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTeachersPayrollSummary(w http.ResponseWriter, r *http.Request, startDate, endDate string) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	roleFilters := parseReportRoleFilters(r)
	ctx := r.Context()

	payrollRows, err := buildTeachersPayrollRows(ctx, q, startDate, endDate, roleFilters)
	if err != nil {
		logs.Log().Error("build teachers payroll rows", zap.Error(err))
		HttpError(w, "Failed to load report data", http.StatusInternalServerError)
		return
	}
	if len(payrollRows) == 0 {
		HttpError(w, "No class data for the selected range.", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("teachers_payroll_%s_%s_%s.xlsx", startDate, endDate, utils.RandomString(8))
	outputPath := filepath.Join("tmp", filename)
	if err := processor.SaveTeachersPayrollReport(payrollRows, outputPath); err != nil {
		logs.Log().Error("save teachers payroll report xlsx", zap.Error(err))
		HttpError(w, "Failed to generate teachers payroll report", http.StatusInternalServerError)
		return
	}

	if err := putReportFile(ctx, filename, outputPath); err != nil {
		logs.Log().Error("upload teachers payroll report to storage", zap.Error(err))
		HttpError(w, "Failed to save teachers payroll report", http.StatusInternalServerError)
		return
	}

	existing, existingErr := dbRO.GetQueries().GetTeachersPayrollReportGeneration(ctx, queries.GetTeachersPayrollReportGenerationParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if existingErr == nil && existing.OutputPath != "" {
		oldBase := reportOutputBasename(existing.OutputPath)
		if oldBase != "" && oldBase != filename {
			_ = storage.Default().Delete(ctx, storage.CategoryReports, oldBase)
			_ = os.Remove(filepath.Join("tmp", oldBase))
		}
	}

	if err := dbRW.GetQueries().UpsertTeachersPayrollReportGeneration(ctx, queries.UpsertTeachersPayrollReportGenerationParams{
		StartDate:   startDate,
		EndDate:     endDate,
		ContentHash: string(constants.ReportGenerationKindTeachersPayroll),
		OutputPath:  filename,
		RecordCount: int64(len(payrollRows)),
	}); err != nil {
		logs.Log().Error("upsert teachers payroll report generation", zap.Error(err))
		HttpError(w, "Failed to save teachers payroll report", http.StatusInternalServerError)
		return
	}

	user := auth.GetUser(ctx)
	insertAuditLogAs(ctx, user, "reports", fmt.Sprintf(
		"generated teachers payroll report (%s to %s): %s",
		startDate, endDate, filename,
	))

	if !serveReportDownload(w, r, filename) {
		HttpError(w, "Failed to download teachers payroll report", http.StatusInternalServerError)
	}
}

type reportRoleFilters struct {
	Teacher bool
	Admin   bool
}

func parseReportRoleFilters(r *http.Request) reportRoleFilters {
	return reportRoleFilters{
		Teacher: parseConnectionCheckboxFilter(r, "roleTeacher"),
		Admin:   parseConnectionCheckboxFilter(r, "roleAdmin"),
	}
}

func reportRoleFilterFlags(filters reportRoleFilters) (int64, int64) {
	teacherFlag := int64(0)
	if filters.Teacher {
		teacherFlag = 1
	}
	adminFlag := int64(0)
	if filters.Admin {
		adminFlag = 1
	}
	return teacherFlag, adminFlag
}

func loadReportRows(ctx context.Context, startDate, endDate, q string, roleFilters reportRoleFilters) ([]frontend.ReportRowData, error) {
	searchParams := reportSearchParams(q, startDate, endDate, roleFilters)

	summaries, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to load report summaries")
	}

	earningsRows, err := dbRO.GetQueries().GetReportTeacherEarnings(ctx, reportEarningsParams(q, startDate, endDate, roleFilters))
	if err != nil {
		return nil, fmt.Errorf("failed to load report earnings")
	}

	earningsByTeacher := map[int64][]reportEarningJSON{}
	for _, row := range earningsRows {
		earningsByTeacher[row.TeacherID] = append(earningsByTeacher[row.TeacherID], reportEarningJSON{
			Currency: row.Currency,
			Total:    sqlNumericToFloat64(row.TotalRate),
		})
	}

	teacherRoleFlag, adminRoleFlag := reportRoleFilterFlags(roleFilters)
	fingerprintRows, err := dbRO.GetQueries().GetClassRecordFingerprintRowsForRange(ctx, queries.GetClassRecordFingerprintRowsForRangeParams{
		Date:    startDate,
		Date_2:  endDate,
		Column3: teacherRoleFlag,
		Column4: adminRoleFlag,
		Column5: teacherRoleFlag,
		Column6: adminRoleFlag,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load report fingerprints")
	}

	hashesByTeacher := map[int64]string{}
	rowsByTeacher := map[int64][]reports.FingerprintRow{}
	for _, row := range fingerprintRows {
		fpRow := fingerprintRowFromRange(row)
		rowsByTeacher[row.TeacherID] = append(rowsByTeacher[row.TeacherID], fpRow)
	}
	for teacherID, fpRows := range rowsByTeacher {
		hashesByTeacher[teacherID] = reports.Fingerprint(fpRows)
	}

	cachedRows, err := dbRO.GetQueries().GetReportGenerationsForRange(ctx, queries.GetReportGenerationsForRangeParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load cached reports")
	}

	cacheByTeacher := map[int64]queries.TblReportGeneration{}
	for _, row := range cachedRows {
		if row.TeacherID.Valid {
			cacheByTeacher[row.TeacherID.Int64] = row
		}
	}

	teacherIDs := make([]int64, len(summaries))
	for i, summary := range summaries {
		teacherIDs[i] = summary.TeacherID
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

	paymentStatuses, err := loadPaymentStatusesForPeriod(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to load payment statuses")
	}

	response := make([]frontend.ReportRowData, 0, len(summaries))
	for _, summary := range summaries {
		item := frontend.ReportRowData{
			TeacherID:   strconv.FormatInt(summary.TeacherID, 10),
			TeacherName: summary.TeacherName,
			TeacherAvatar: avatarWithTeacherRoles(
				buildReportSummaryAvatarProps(summary),
				rolesMap[summary.TeacherID],
				teacherStatusFromMap(statusMap, summary.TeacherID),
			),
			ConductedClasses: sqlNumericToInt64(summary.ConductedClasses),
			CancelledClasses: sqlNumericToInt64(summary.CancelledClasses),
			TotalClasses:     summary.TotalClasses,
			Earnings:         reportEarningsToFrontend(earningsByTeacher[summary.TeacherID]),
			PaymentStatus:    paymentStatuses[summary.TeacherID],
		}
		if item.HasPaymentStatus() {
			item.SendPaymentDisabledTooltip = paymentSentDisabledTooltip(startDate, endDate)
		}

		currentHash := hashesByTeacher[summary.TeacherID]
		if cache, ok := cacheByTeacher[summary.TeacherID]; ok && cache.ContentHash == currentHash {
			if filename, ok := reportCacheAvailable(ctx, cache.OutputPath); ok {
				item.DownloadReady = true
				item.Filename = filename
			}
		}

		response = append(response, item)
	}
	return response, nil
}

func loadReportRow(ctx context.Context, teacherID int64, startDate, endDate string) (frontend.ReportRowData, error) {
	rows, err := loadReportRows(ctx, startDate, endDate, "", reportRoleFilters{})
	if err != nil {
		return frontend.ReportRowData{}, err
	}
	id := strconv.FormatInt(teacherID, 10)
	for _, row := range rows {
		if row.TeacherID == id {
			return row, nil
		}
	}
	return frontend.ReportRowData{}, ErrTeacherNotFoundInReportSummaries
}

func aggregateReportEarnings(rows []frontend.ReportRowData) []frontend.CurrencyTotal {
	totalsByCurrency := map[string]float64{}
	for _, row := range rows {
		for _, earning := range row.Earnings {
			if earning.Total == 0 {
				continue
			}
			totalsByCurrency[earning.Currency] += earning.Total
		}
	}
	if len(totalsByCurrency) == 0 {
		return nil
	}
	out := make([]frontend.CurrencyTotal, 0, len(totalsByCurrency))
	for _, currency := range constants.CurrencyCodes {
		total, ok := totalsByCurrency[currency]
		if !ok || total == 0 {
			continue
		}
		out = append(out, frontend.CurrencyTotal{
			Currency: currency,
			Total:    total,
		})
	}
	return out
}

func reportEarningsToFrontend(earnings []reportEarningJSON) []frontend.ReportEarningData {
	if len(earnings) == 0 {
		return []frontend.ReportEarningData{}
	}
	out := make([]frontend.ReportEarningData, 0, len(earnings))
	for _, e := range earnings {
		out = append(out, frontend.ReportEarningData{
			Currency: e.Currency,
			Total:    e.Total,
		})
	}
	return out
}

func renderReportGenerateRow(w http.ResponseWriter, r *http.Request, teacherID int64, startDate, endDate string, skipped bool) {
	row, err := loadReportRow(r.Context(), teacherID, startDate, endDate)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if skipped {
		w.Header().Set(headerHXTrigger, `{"showSuccessBanner":"Report is up to date; skipped generation"}`)
	}
	writeHTML(w)
	if err := frontend.ReportTableRow(row, startDate, endDate, false).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportView(w http.ResponseWriter, r *http.Request, teacherID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		HttpError(w, MsgTeacherNotFound, http.StatusNotFound)
		return
	}

	records, err := dbRO.GetQueries().GetTeacherReportClassRecords(ctx, queries.GetTeacherReportClassRecordsParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		HttpError(w, "Failed to load class records", http.StatusInternalServerError)
		return
	}

	statusCounts, err := dbRO.GetQueries().CountClassRecordsByStatusAndDateRange(ctx, queries.CountClassRecordsByStatusAndDateRangeParams{
		Date:      startDate,
		Date_2:    endDate,
		Column3:   teacherID,
		TeacherID: teacherID,
	})
	if err != nil {
		HttpError(w, "Failed to load class summary", http.StatusInternalServerError)
		return
	}

	var conducted, rescheduled, cancelled int64
	for _, row := range statusCounts {
		switch row.Status {
		case "conducted":
			conducted = row.Count
		case "rescheduled":
			rescheduled = row.Count
		case "cancelled":
			cancelled = row.Count
		}
	}

	classes := make([]frontend.ReportClassItem, 0, len(records))
	for _, record := range records {
		classes = append(classes, frontend.ReportClassItem{
			Date:            record.Date,
			StudentName:     studentDisplayName(record.StudentName, record.StudentNumber, sql.NullString{}),
			Rate:            utils.FormatCurrency(record.Rate, record.Currency),
			DurationMinutes: record.DurationMinutes,
			TimeRange:       formatReportTimeRange(record.StartTime, record.EndTime),
			Status:          constants.ClassStatus(record.Status),
		})
	}

	earningsRows, err := dbRO.GetQueries().SumConductedRateByCurrencyAndDateRange(ctx, queries.SumConductedRateByCurrencyAndDateRangeParams{
		Date:      startDate,
		Date_2:    endDate,
		Column3:   teacherID,
		TeacherID: teacherID,
	})
	if err != nil {
		HttpError(w, "Failed to load report earnings", http.StatusInternalServerError)
		return
	}

	totalEarnings := make([]frontend.CurrencyTotal, 0, len(earningsRows))
	for _, row := range earningsRows {
		totalEarnings = append(totalEarnings, frontend.CurrencyTotal{
			Currency: row.Currency,
			Total:    sqlNumericToFloat64(row.TotalRate),
		})
	}

	writeHTML(w)
	teacherName := utils.ComposePersonName(profile.FirstName, profile.MiddleName, profile.LastName)
	teacherRoles, err := loadTeacherRoles(ctx, teacherID)
	if err != nil {
		HttpError(w, MsgFailedToLoadTeacherRoles, http.StatusInternalServerError)
		return
	}
	frontend.ReportViewModal(frontend.ReportViewData{
		TeacherID:        strconv.FormatInt(teacherID, 10),
		TeacherName:      teacherName,
		CutoffLabel:      formatReportCutoffLabel(startDate, endDate),
		TotalEarnings:    totalEarnings,
		ConductedCount:   conducted,
		RescheduledCount: rescheduled,
		CancelledCount:   cancelled,
		Classes:          classes,
		Avatar:           avatarWithTeacherRoles(buildReportTeacherAvatarProps(teacherID, profile), teacherRoles, constants.TeacherStatus(profile.Status)),
	}).Render(ctx, w)
}

func handleReportGenerate(w http.ResponseWriter, r *http.Request, teacherID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, "invalid request")
		return
	}

	startDate := strings.TrimSpace(r.FormValue("startDate"))
	endDate := strings.TrimSpace(r.FormValue("endDate"))
	if startDate == "" || endDate == "" {
		sendErrorLog(w, "start and end dates are required")
		return
	}
	if startDate > endDate {
		sendErrorLog(w, "end date must be after start date")
		return
	}

	ctx := r.Context()
	filename, _, fromCache, err := generateTeacherReportFile(ctx, teacherID, startDate, endDate)
	if err != nil {
		if errors.Is(err, ErrReportNoRecords) {
			sendErrorLog(w, err.Error())
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorLog(w, MsgTeacherNotFound)
			return
		}
		sendErrorLog(w, "failed to generate report")
		return
	}
	if fromCache {
		renderReportGenerateRow(w, r, teacherID, startDate, endDate, true)
		return
	}
	_ = filename
	renderReportGenerateRow(w, r, teacherID, startDate, endDate, false)
}

func generateTeacherReportFile(ctx context.Context, teacherID int64, startDate, endDate string) (filename string, recordCount int64, fromCache bool, err error) {
	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		return "", 0, false, err
	}

	fpRows, err := dbRO.GetQueries().GetClassRecordFingerprintRows(ctx, queries.GetClassRecordFingerprintRowsParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		return "", 0, false, err
	}
	currentHash := reports.Fingerprint(fingerprintRowsFromFingerprintQuery(fpRows))

	records, err := dbRO.GetQueries().GetTeacherReportClassRecords(ctx, queries.GetTeacherReportClassRecordsParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		return "", 0, false, err
	}
	if len(records) == 0 {
		return "", 0, false, ErrReportNoRecords
	}

	cache, cacheErr := dbRO.GetQueries().GetReportGeneration(ctx, queries.GetReportGenerationParams{
		TeacherID: sql.NullInt64{Int64: teacherID, Valid: true},
		StartDate: startDate,
		EndDate:   endDate,
	})
	if cacheErr == nil && cache.ContentHash == currentHash {
		if cachedName, ok := reportCacheAvailable(ctx, cache.OutputPath); ok {
			return cachedName, int64(len(records)), true, nil
		}
	}

	processorRecords := classRecordsToProcessor(records)
	teacherName := utils.ComposePersonName(profile.FirstName, profile.MiddleName, profile.LastName)
	safeName := utils.SanitizeFilename(teacherName)
	filename = fmt.Sprintf("%s_report_%s.xlsx", safeName, utils.RandomString(8))
	outputPath := filepath.Join("tmp", filename)

	colIndices := processor.ColumnIndices{
		Name:      0,
		Duration:  0,
		Rate:      0,
		Status:    0,
		StartTime: 0,
		EndTime:   0,
		Link:      -1,
	}
	if err := processor.SaveReportRecords(processorRecords, outputPath, colIndices, teacherName); err != nil {
		logs.Log().Error("save report xlsx", zap.Error(err))
		return "", 0, false, err
	}

	if err := putReportFile(ctx, filename, outputPath); err != nil {
		logs.Log().Error("upload report to storage", zap.Error(err))
		return "", 0, false, err
	}

	if cacheErr == nil && cache.OutputPath != "" {
		oldBase := reportOutputBasename(cache.OutputPath)
		if oldBase != "" && oldBase != filename {
			_ = storage.Default().Delete(ctx, storage.CategoryReports, oldBase)
			_ = os.Remove(filepath.Join("tmp", oldBase))
		}
	}

	if err := dbRW.GetQueries().UpsertReportGeneration(ctx, queries.UpsertReportGenerationParams{
		TeacherID:   sql.NullInt64{Int64: teacherID, Valid: true},
		StartDate:   startDate,
		EndDate:     endDate,
		ContentHash: currentHash,
		OutputPath:  filename,
		RecordCount: int64(len(records)),
	}); err != nil {
		logs.Log().Error("upsert report generation", zap.Error(err))
		return "", 0, false, err
	}

	user := auth.GetUser(ctx)
	insertAuditLogAs(ctx, user, "reports", fmt.Sprintf(
		"generated report for teacher '%s' (%s to %s): %d records, %s",
		teacherName, startDate, endDate, len(records), filename,
	))

	return filename, int64(len(records)), false, nil
}

func reportSearchParams(q, startDate, endDate string, roleFilters reportRoleFilters) queries.GetReportTeacherSummariesParams {
	qNull := sql.NullString{String: q, Valid: q != ""}
	teacherRoleFlag, adminRoleFlag := reportRoleFilterFlags(roleFilters)
	return queries.GetReportTeacherSummariesParams{
		Date:     startDate,
		Date_2:   endDate,
		Column3:  q,
		Column4:  qNull,
		Date_3:   startDate,
		Date_4:   endDate,
		Column7:  qNull,
		Column8:  teacherRoleFlag,
		Column9:  adminRoleFlag,
		Column10: teacherRoleFlag,
		Column11: adminRoleFlag,
	}
}

func reportSummaryParams(q, startDate, endDate string, roleFilters reportRoleFilters) queries.GetReportSummaryRowsParams {
	qNull := sql.NullString{String: q, Valid: q != ""}
	teacherRoleFlag, adminRoleFlag := reportRoleFilterFlags(roleFilters)
	return queries.GetReportSummaryRowsParams{
		Date:     startDate,
		Date_2:   endDate,
		Column3:  q,
		Column4:  qNull,
		Date_3:   startDate,
		Date_4:   endDate,
		Column7:  qNull,
		Column8:  teacherRoleFlag,
		Column9:  adminRoleFlag,
		Column10: teacherRoleFlag,
		Column11: adminRoleFlag,
	}
}

func requireReportDateRange(r *http.Request) (string, string, error) {
	return parseListDateRange(r)
}

func fingerprintRowFromRange(row queries.GetClassRecordFingerprintRowsForRangeRow) reports.FingerprintRow {
	return reports.FingerprintRow{
		ID:              row.ID,
		StudentID:       row.StudentID,
		Date:            row.Date,
		StartTime:       row.StartTime,
		EndTime:         row.EndTime,
		DurationMinutes: row.DurationMinutes,
		Rate:            row.Rate,
		Currency:        row.Currency,
		Status:          row.Status,
		UpdatedAt:       row.UpdatedAt,
	}
}

func reportEarningsParams(q, startDate, endDate string, roleFilters reportRoleFilters) queries.GetReportTeacherEarningsParams {
	qNull := sql.NullString{String: q, Valid: q != ""}
	teacherRoleFlag, adminRoleFlag := reportRoleFilterFlags(roleFilters)
	return queries.GetReportTeacherEarningsParams{
		Date:     startDate,
		Date_2:   endDate,
		Column3:  q,
		Column4:  qNull,
		Date_3:   startDate,
		Date_4:   endDate,
		Column7:  qNull,
		Column8:  teacherRoleFlag,
		Column9:  adminRoleFlag,
		Column10: teacherRoleFlag,
		Column11: adminRoleFlag,
	}
}

func fingerprintRowsFromFingerprintQuery(rows []queries.GetClassRecordFingerprintRowsRow) []reports.FingerprintRow {
	out := make([]reports.FingerprintRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, reports.FingerprintRow{
			ID:              row.ID,
			StudentID:       row.StudentID,
			Date:            row.Date,
			StartTime:       row.StartTime,
			EndTime:         row.EndTime,
			DurationMinutes: row.DurationMinutes,
			Rate:            row.Rate,
			Currency:        row.Currency,
			Status:          row.Status,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out
}

func classRecordsToProcessor(records []queries.GetTeacherReportClassRecordsRow) []processor.ClassRecord {
	out := make([]processor.ClassRecord, 0, len(records))
	for _, record := range records {
		item := processor.ClassRecord{
			Name:              studentDisplayName(record.StudentName, record.StudentNumber, sql.NullString{}),
			DurationInMinutes: int(record.DurationMinutes),
			Rate:              record.Rate,
			Status:            record.Status,
			Date:              record.Date,
		}
		if record.StartTime.Valid {
			if t, err := time.Parse("15:04", record.StartTime.String); err == nil {
				item.StartTime = &t
			}
		}
		if record.EndTime.Valid {
			if t, err := time.Parse("15:04", record.EndTime.String); err == nil {
				item.EndTime = &t
			}
		}
		out = append(out, item)
	}
	return out
}

type summaryStudentData struct {
	name     string
	classes  []processor.SummaryClassRow
	total    float64
	currency string
	hasTotal bool
}

func buildSummaryTeacherSheets(summaries []queries.GetReportTeacherSummariesRow, rows []queries.GetReportSummaryRowsRow) []processor.SummaryTeacherSheet {
	teacherStudents := map[int64]map[int64]*summaryStudentData{}
	teacherStudentOrder := map[int64][]int64{}

	for _, row := range rows {
		if row.IsTrialClass != 0 {
			continue
		}
		if _, ok := teacherStudents[row.TeacherID]; !ok {
			teacherStudents[row.TeacherID] = map[int64]*summaryStudentData{}
			teacherStudentOrder[row.TeacherID] = []int64{}
		}
		if _, ok := teacherStudents[row.TeacherID][row.StudentID]; !ok {
			teacherStudents[row.TeacherID][row.StudentID] = &summaryStudentData{name: studentDisplayName(row.StudentName, row.StudentNumber, sql.NullString{})}
			teacherStudentOrder[row.TeacherID] = append(teacherStudentOrder[row.TeacherID], row.StudentID)
		}
		student := teacherStudents[row.TeacherID][row.StudentID]
		classRow := summaryClassRowFromRecord(row)
		student.classes = append(student.classes, classRow)
		if classRow.HasRate {
			student.total += classRow.RateValue
			student.currency = row.ParentCurrency.String
			student.hasTotal = true
		}
	}

	sheets := make([]processor.SummaryTeacherSheet, 0, len(summaries))
	for _, summary := range summaries {
		sheet := processor.SummaryTeacherSheet{TeacherName: summary.TeacherName}
		for _, studentID := range teacherStudentOrder[summary.TeacherID] {
			student := teacherStudents[summary.TeacherID][studentID]
			sheet.Students = append(sheet.Students, processor.SummaryStudentBlock{
				StudentName: student.name,
				Classes:     student.classes,
				Total:       student.total,
				Currency:    student.currency,
				HasTotal:    student.hasTotal,
			})
		}
		sheets = append(sheets, sheet)
	}
	return sheets
}

func summaryClassRowFromRecord(row queries.GetReportSummaryRowsRow) processor.SummaryClassRow {
	status := constants.ClassStatus(row.Status)
	parentDisplay, parentValue, hasParent := effectiveSummaryParentRate(
		row.ParentRate, row.ParentCurrency, status, row.Rate, row.TeacherRatePerClass,
		row.DurationMinutes, row.TeacherRateDurationMinutes,
	)
	return processor.SummaryClassRow{
		Status:           status,
		StatusDisplay:    status.Label(),
		DateDisplay:      formatReportSummaryClassDate(row.Date),
		DurationDisplay:  formatSummaryClassDurationDisplay(row.DurationMinutes),
		RateDisplay:      parentDisplay,
		RateValue:        parentValue,
		HasRate:          hasParent,
		ClassRateDisplay: formatSummaryClassRateDisplay(row.Rate, row.Currency),
	}
}

func formatSummaryClassDurationDisplay(durationMinutes int64) string {
	if durationMinutes <= 0 {
		return constants.DisplayNotSet
	}
	return fmt.Sprintf("%d", durationMinutes)
}

func effectiveSummaryParentRate(
	parentRate sql.NullFloat64,
	parentCurrency sql.NullString,
	status constants.ClassStatus,
	classRate, teacherRate float64,
	classDurationMinutes, teacherRateDurationMinutes int64,
) (string, float64, bool) {
	if !parentRate.Valid || !parentCurrency.Valid || parentCurrency.String == "" {
		return constants.DisplayNotSet, 0, false
	}
	amount := billing.EffectiveParentBillAmount(
		status, parentRate.Float64, classRate, teacherRate,
		classDurationMinutes, teacherRateDurationMinutes,
	)
	currency := parentCurrency.String
	return utils.FormatCurrencyAmountFixed(amount, currency) + " " + currency, amount, true
}

func formatSummaryClassRateDisplay(rate float64, currency string) string {
	if currency == "" {
		return constants.DisplayNotSet
	}
	return utils.FormatCurrencyAmountFixed(rate, currency) + " " + currency
}

func formatReportSummaryClassDate(dateStr string) string {
	t, err := time.ParseInLocation(constants.DateLayout, dateStr, constants.LocationPHT)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2")
}

func buildTeachersPayrollRows(ctx context.Context, q, startDate, endDate string, roleFilters reportRoleFilters) ([]processor.TeachersPayrollRow, error) {
	presets := utils.CutoffPresetsContainedInRange(startDate, endDate)
	if len(presets) == 0 {
		return nil, nil
	}

	cutoffsByTeacher := map[int64][]processor.TeachersPayrollRow{}
	for _, preset := range presets {
		cutoffStart, cutoffEnd := utils.CutoffDatesFromPreset(preset)
		if cutoffStart == "" || cutoffEnd == "" {
			continue
		}
		cutoffLabel := formatPayrollCutoffPeriodLabel(cutoffStart, cutoffEnd)

		summaries, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, reportSearchParams(q, cutoffStart, cutoffEnd, roleFilters))
		if err != nil {
			return nil, err
		}
		earningsRows, err := dbRO.GetQueries().GetReportTeacherEarnings(ctx, reportEarningsParams(q, cutoffStart, cutoffEnd, roleFilters))
		if err != nil {
			return nil, err
		}
		earningsByTeacher := map[int64][]reportEarningJSON{}
		for _, row := range earningsRows {
			earningsByTeacher[row.TeacherID] = append(earningsByTeacher[row.TeacherID], reportEarningJSON{
				Currency: row.Currency,
				Total:    sqlNumericToFloat64(row.TotalRate),
			})
		}

		for _, summary := range summaries {
			if summary.TotalClasses == 0 {
				continue
			}
			fullName := utils.ComposePersonName(summary.TeacherFirstName, summary.TeacherMiddleName, summary.TeacherLastName)
			cutoffsByTeacher[summary.TeacherID] = append(cutoffsByTeacher[summary.TeacherID], processor.TeachersPayrollRow{
				TeacherID:       summary.TeacherID,
				TeacherFullName: fullName,
				CutoffLabel:     cutoffLabel,
				TotalClasses:    summary.TotalClasses,
				EarningsDisplay: formatTeachersPayrollEarnings(earningsByTeacher[summary.TeacherID]),
			})
		}
	}

	orderSummaries, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, reportSearchParams(q, startDate, endDate, roleFilters))
	if err != nil {
		return nil, err
	}

	var rows []processor.TeachersPayrollRow
	for _, summary := range orderSummaries {
		teacherRows := cutoffsByTeacher[summary.TeacherID]
		if len(teacherRows) == 0 {
			continue
		}
		rows = append(rows, teacherRows...)
	}
	return rows, nil
}

func formatTeachersPayrollEarnings(earnings []reportEarningJSON) string {
	if len(earnings) == 0 {
		return constants.DisplayNotSet
	}
	parts := make([]string, 0, len(earnings))
	for _, item := range earnings {
		if item.Total == 0 {
			continue
		}
		parts = append(parts, utils.FormatCurrency(item.Total, item.Currency))
	}
	if len(parts) == 0 {
		return constants.DisplayNotSet
	}
	return strings.Join(parts, ", ")
}

func formatPayrollCutoffPeriodLabel(startDate, endDate string) string {
	start, err := time.ParseInLocation(constants.DateLayout, startDate, constants.LocationPHT)
	if err != nil {
		return startDate + " - " + endDate
	}
	end, err := time.ParseInLocation(constants.DateLayout, endDate, constants.LocationPHT)
	if err != nil {
		return startDate + " - " + endDate
	}
	if start.Month() == end.Month() && start.Year() == end.Year() {
		return fmt.Sprintf("%s %d - %d", start.Format("January"), start.Day(), end.Day())
	}
	return fmt.Sprintf("%s %d - %s %d", start.Format("January"), start.Day(), end.Format("January"), end.Day())
}

func formatReportCutoffLabel(startDate, endDate string) string {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return startDate + " – " + endDate
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return startDate + " – " + endDate
	}
	return fmt.Sprintf("%s – %s", start.Format("2 Jan 2006"), end.Format("2 Jan 2006"))
}

func formatReportTimeRange(startTime, endTime sql.NullString) string {
	if startTime.Valid && endTime.Valid {
		return startTime.String + " – " + endTime.String
	}
	if startTime.Valid {
		return startTime.String
	}
	if endTime.Valid {
		return endTime.String
	}
	return constants.DisplayNotSet
}

func buildReportTeacherAvatarProps(teacherID int64, row queries.GetTeacherProfileByIDRow) frontend.AvatarProps {
	hasPicture := row.ProfilePicture.Valid && row.ProfilePicture.String != ""
	assignedColor := row.AssignedColor
	if assignedColor == "" {
		assignedColor = constants.DefaultTeacherAssignedColor
	}
	displayName := utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName)
	return frontend.AvatarProps{
		Size:          "lg",
		Initials:      utils.PersonInitials(row.FirstName, row.MiddleName, row.LastName, displayName),
		AssignedColor: assignedColor,
		PictureURL:    teacherPictureURL(teacherID, hasPicture),
		HasPicture:    hasPicture,
		Alt:           displayName + " avatar",
	}
}

func buildParentsSummaryViewData(summaries []queries.GetReportTeacherSummariesRow, sheets []processor.SummaryTeacherSheet, startDate, endDate string) frontend.ReportSummaryViewData {
	teachers := make([]frontend.ParentsSummaryTeacherBlock, 0, len(sheets))
	for i, sheet := range sheets {
		var avatar frontend.AvatarProps
		if i < len(summaries) {
			avatar = buildReportSummaryAvatarProps(summaries[i])
		}
		students := make([]frontend.ParentsSummaryStudentBlock, 0, len(sheet.Students))
		for _, student := range sheet.Students {
			classes := make([]frontend.ParentsSummaryClassRow, 0, len(student.Classes))
			for _, classRow := range student.Classes {
				classes = append(classes, frontend.ParentsSummaryClassRow{
					Status:     classRow.Status,
					Date:       classRow.DateDisplay,
					Duration:   classRow.DurationDisplay,
					ParentRate: classRow.RateDisplay,
					ClassRate:  classRow.ClassRateDisplay,
				})
			}
			block := frontend.ParentsSummaryStudentBlock{
				StudentName: student.StudentName,
				Classes:     classes,
				HasTotal:    student.HasTotal,
			}
			if student.HasTotal {
				block.TotalLabel = fmt.Sprintf("total in %s", strings.ToLower(student.Currency))
				block.TotalDisplay = utils.FormatCurrencyAmountFixed(student.Total, student.Currency) + " " + student.Currency
			}
			students = append(students, block)
		}
		teachers = append(teachers, frontend.ParentsSummaryTeacherBlock{
			TeacherName: sheet.TeacherName,
			Avatar:      avatar,
			Students:    students,
		})
	}
	return frontend.ReportSummaryViewData{
		Kind:        constants.SummaryReportKindParents,
		Title:       "Parents summary",
		CutoffLabel: formatReportCutoffLabel(startDate, endDate),
		Parents:     teachers,
	}
}

func buildTeachersPayrollSummaryViewData(rows []processor.TeachersPayrollRow, avatarByTeacher map[int64]frontend.AvatarProps, startDate, endDate string) frontend.ReportSummaryViewData {
	viewRows := make([]frontend.TeachersPayrollSummaryRow, 0, len(rows))
	for _, row := range rows {
		avatar := avatarByTeacher[row.TeacherID]
		viewRows = append(viewRows, frontend.TeachersPayrollSummaryRow{
			TeacherName:     row.TeacherFullName,
			TeacherAvatar:   avatar,
			CutoffLabel:     row.CutoffLabel,
			TotalClasses:    row.TotalClasses,
			EarningsDisplay: row.EarningsDisplay,
		})
	}
	return frontend.ReportSummaryViewData{
		Kind:        constants.SummaryReportKindTeachersPayroll,
		Title:       "Teachers payroll",
		CutoffLabel: formatReportCutoffLabel(startDate, endDate),
		Payroll:     viewRows,
	}
}

func buildReportSummaryAvatarProps(summary queries.GetReportTeacherSummariesRow) frontend.AvatarProps {
	hasPicture := summary.TeacherProfilePicture.Valid && summary.TeacherProfilePicture.String != ""
	assignedColor := summary.TeacherAssignedColor
	if assignedColor == "" {
		assignedColor = constants.DefaultTeacherAssignedColor
	}
	return frontend.AvatarProps{
		Size:          "sm",
		Initials:      utils.PersonInitials(summary.TeacherFirstName, summary.TeacherMiddleName, summary.TeacherLastName, summary.TeacherName),
		AssignedColor: assignedColor,
		PictureURL:    teacherPictureURL(summary.TeacherID, hasPicture),
		HasPicture:    hasPicture,
		Alt:           summary.TeacherName + " avatar",
	}
}

func sqlNumericToInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		parsed, err := strconv.ParseInt(string(v), 10, 64)
		if err == nil {
			return parsed
		}
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func sqlNumericToFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case []byte:
		parsed, err := strconv.ParseFloat(string(v), 64)
		if err == nil {
			return parsed
		}
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func auditReportDownload(ctx context.Context, filename string) {
	base := reportOutputBasename(filename)
	candidates := []string{base, filepath.Join("tmp", base)}
	var row queries.GetReportGenerationByOutputBasenameRow
	var err error
	for _, outputPath := range candidates {
		row, err = dbRO.GetQueries().GetReportGenerationByOutputBasename(ctx, outputPath)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}

	user := auth.GetUser(ctx)
	insertAuditLogAs(ctx, user, "reports", fmt.Sprintf(
		"downloaded report for teacher '%s' (%s to %s): %s",
		fmt.Sprint(row.TeacherName), row.StartDate, row.EndDate, filename,
	))
}
