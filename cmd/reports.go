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
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/processor"
	"zion-english/internal/reports"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

type reportEarningJSON struct {
	Currency string  `json:"currency"`
	Total    float64 `json:"total"`
}

func handleReportsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "reports", "/view"); ok {
		handleReportView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "reports", "/generate"); ok {
		handleReportGenerate(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if err := frontend.Reports().Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportsPartial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := loadReportRows(r.Context(), startDate, endDate, q)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	emptyMsg := "No teachers found."
	if startDate == "" || endDate == "" {
		emptyMsg = "Select a date range."
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ReportsTableBody(rows, startDate, endDate, emptyMsg).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadReportRows(ctx context.Context, startDate, endDate, q string) ([]frontend.ReportRowData, error) {
	searchParams := reportSearchParams(q, startDate, endDate)

	summaries, err := dbRO.GetQueries().GetReportTeacherSummaries(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to load report summaries")
	}

	earningsRows, err := dbRO.GetQueries().GetReportTeacherEarnings(ctx, reportEarningsParams(q, startDate, endDate))
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

	fingerprintRows, err := dbRO.GetQueries().GetClassRecordFingerprintRowsForRange(ctx, queries.GetClassRecordFingerprintRowsForRangeParams{
		Date:   startDate,
		Date_2: endDate,
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
		cacheByTeacher[row.TeacherID] = row
	}

	teacherIDs := make([]int64, len(summaries))
	for i, summary := range summaries {
		teacherIDs[i] = summary.TeacherID
	}
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to load teacher roles")
	}

	response := make([]frontend.ReportRowData, 0, len(summaries))
	for _, summary := range summaries {
		item := frontend.ReportRowData{
			TeacherID:        strconv.FormatInt(summary.TeacherID, 10),
			TeacherName:      summary.TeacherName,
			TeacherAvatar: avatarWithTeacherRoles(
				buildReportSummaryAvatarProps(summary),
				rolesMap[summary.TeacherID],
			),
			ConductedClasses: sqlNumericToInt64(summary.ConductedClasses),
			TotalClasses:     summary.TotalClasses,
			Earnings:         reportEarningsToFrontend(earningsByTeacher[summary.TeacherID]),
		}

		currentHash := hashesByTeacher[summary.TeacherID]
		if cache, ok := cacheByTeacher[summary.TeacherID]; ok && cache.ContentHash == currentHash {
			if filename, ok := reportCacheFilename(cache.OutputPath); ok {
				item.DownloadReady = true
				item.Filename = filename
			}
		}

		response = append(response, item)
	}
	return response, nil
}

func loadReportRow(ctx context.Context, teacherID int64, startDate, endDate string) (frontend.ReportRowData, error) {
	rows, err := loadReportRows(ctx, startDate, endDate, "")
	if err != nil {
		return frontend.ReportRowData{}, err
	}
	id := strconv.FormatInt(teacherID, 10)
	for _, row := range rows {
		if row.TeacherID == id {
			return row, nil
		}
	}
	return frontend.ReportRowData{}, errors.New("teacher not found in report summaries")
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
		w.Header().Set("HX-Trigger", `{"showSuccessBanner":"Report is up to date — skipped generation"}`)
	}
	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ReportTableRow(row, startDate, endDate).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleReportView(w http.ResponseWriter, r *http.Request, teacherID int64) {
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
	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		HttpError(w, "Teacher not found", http.StatusNotFound)
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
			Date:        record.Date,
			StudentName: record.StudentName,
			Rate:        fmt.Sprintf("%.2f %s", record.Rate, record.Currency),
			TimeRange:   formatReportTimeRange(record.StartTime, record.EndTime),
			Status:      record.Status,
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

	w.Header().Set("Content-Type", "text/html")
	teacherName := utils.ComposePersonName(profile.FirstName, profile.MiddleName, profile.LastName)
	teacherRoles, err := loadTeacherRoles(ctx, teacherID)
	if err != nil {
		HttpError(w, "Failed to load teacher roles", http.StatusInternalServerError)
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
		Avatar:           avatarWithTeacherRoles(buildReportTeacherAvatarProps(teacherID, profile), teacherRoles),
	}).Render(ctx, w)
}

func handleReportGenerate(w http.ResponseWriter, r *http.Request, teacherID int64) {
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
	if startDate == "" || endDate == "" {
		sendErrorLog(w, "start and end dates are required")
		return
	}
	if startDate > endDate {
		sendErrorLog(w, "end date must be after start date")
		return
	}

	ctx := r.Context()
	profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		sendErrorLog(w, "teacher not found")
		return
	}

	fpRows, err := dbRO.GetQueries().GetClassRecordFingerprintRows(ctx, queries.GetClassRecordFingerprintRowsParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		sendErrorLog(w, "failed to load class records")
		return
	}
	currentHash := reports.Fingerprint(fingerprintRowsFromFingerprintQuery(fpRows))

	records, err := dbRO.GetQueries().GetTeacherReportClassRecords(ctx, queries.GetTeacherReportClassRecordsParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		sendErrorLog(w, "failed to load class records")
		return
	}
	if len(records) == 0 {
		sendErrorLog(w, "no records found")
		return
	}

	cache, cacheErr := dbRO.GetQueries().GetReportGeneration(ctx, queries.GetReportGenerationParams{
		TeacherID: teacherID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if cacheErr == nil && cache.ContentHash == currentHash {
		if _, ok := reportCacheFilename(cache.OutputPath); ok {
			renderReportGenerateRow(w, r, teacherID, startDate, endDate, true)
			return
		}
	}

	processorRecords := classRecordsToProcessor(records)
	teacherName := utils.ComposePersonName(profile.FirstName, profile.MiddleName, profile.LastName)
	safeName := utils.SanitizeFilename(teacherName)
	filename := fmt.Sprintf("%s_report_%s.xlsx", safeName, utils.RandomString(8))
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
		sendErrorLog(w, "failed to generate report")
		return
	}

	if cacheErr == nil && cache.OutputPath != "" && cache.OutputPath != outputPath {
		_ = os.Remove(cache.OutputPath)
	}

	if err := dbRW.GetQueries().UpsertReportGeneration(ctx, queries.UpsertReportGenerationParams{
		TeacherID:   teacherID,
		StartDate:   startDate,
		EndDate:     endDate,
		ContentHash: currentHash,
		OutputPath:  outputPath,
		RecordCount: int64(len(records)),
	}); err != nil {
		logs.Log().Error("upsert report generation", zap.Error(err))
		sendErrorLog(w, "failed to save report cache")
		return
	}

	user := auth.GetUser(ctx)
	insertAuditLogAs(ctx, user, "reports", fmt.Sprintf(
		"generated report for teacher '%s' (%s to %s): %d records, %s",
		teacherName, startDate, endDate, len(records), filename,
	))

	renderReportGenerateRow(w, r, teacherID, startDate, endDate, false)
}

func reportSearchParams(q, startDate, endDate string) queries.GetReportTeacherSummariesParams {
	qNull := sql.NullString{String: q, Valid: q != ""}
	return queries.GetReportTeacherSummariesParams{
		Date:    startDate,
		Date_2:  endDate,
		Column3: q,
		Column4: qNull,
		Date_3:  startDate,
		Date_4:  endDate,
		Column7: qNull,
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

func reportEarningsParams(q, startDate, endDate string) queries.GetReportTeacherEarningsParams {
	qNull := sql.NullString{String: q, Valid: q != ""}
	return queries.GetReportTeacherEarningsParams{
		Date:    startDate,
		Date_2:  endDate,
		Column3: q,
		Column4: qNull,
		Date_3:  startDate,
		Date_4:  endDate,
		Column7: qNull,
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
			Name:              record.StudentName,
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

func reportCacheFilename(outputPath string) (string, bool) {
	if outputPath == "" {
		return "", false
	}
	cleanPath := filepath.Clean(outputPath)
	fullPath := filepath.Join("tmp", filepath.Base(cleanPath))
	if !strings.HasPrefix(fullPath, filepath.Clean("tmp")+string(os.PathSeparator)) {
		return "", false
	}
	if _, err := os.Stat(fullPath); err != nil {
		return "", false
	}
	return filepath.Base(fullPath), true
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
	return "-"
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
	outputPath := filepath.Join("tmp", filepath.Base(filename))
	row, err := dbRO.GetQueries().GetReportGenerationByOutputBasename(ctx, outputPath)
	if err != nil {
		return
	}

	user := auth.GetUser(ctx)
	insertAuditLogAs(ctx, user, "reports", fmt.Sprintf(
		"downloaded report for teacher '%s' (%s to %s): %s",
		row.TeacherName, row.StartDate, row.EndDate, filename,
	))
}
