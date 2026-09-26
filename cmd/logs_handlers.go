package cmd

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type logFilters struct {
	module    string
	message   string
	startDate string
	endDate   string
}

func parseLogFilters(r *http.Request) logFilters {
	return logFilters{
		module:    r.URL.Query().Get("module"),
		message:   r.URL.Query().Get("q"),
		startDate: r.URL.Query().Get("startDate"),
		endDate:   r.URL.Query().Get("endDate"),
	}
}

func countAllLogsParams(f logFilters) queries.CountAllLogsFilteredParams {
	return queries.CountAllLogsFilteredParams{
		Column1:     f.module,
		Module:      f.module,
		Column3:     f.message,
		Column4:     sql.NullString{String: f.message, Valid: true},
		Column5:     f.startDate,
		CreatedAt:   f.startDate,
		Column7:     f.endDate,
		CreatedAt_2: f.endDate,
	}
}

func countLogsByTeacherParams(teacherID int64, f logFilters) queries.CountLogsByCreatedByFilteredParams {
	return queries.CountLogsByCreatedByFilteredParams{
		CreatedBy:   sql.NullInt64{Int64: teacherID, Valid: true},
		Column2:     f.module,
		Module:      f.module,
		Column4:     f.message,
		Column5:     sql.NullString{String: f.message, Valid: true},
		Column6:     f.startDate,
		CreatedAt:   f.startDate,
		Column8:     f.endDate,
		CreatedAt_2: f.endDate,
	}
}

func getAllLogsParams(f logFilters, limit, offset int64) queries.GetAllLogsFilteredParams {
	return queries.GetAllLogsFilteredParams{
		Column1:     f.module,
		Module:      f.module,
		Column3:     f.message,
		Column4:     sql.NullString{String: f.message, Valid: true},
		Column5:     f.startDate,
		CreatedAt:   f.startDate,
		Column7:     f.endDate,
		CreatedAt_2: f.endDate,
		Limit:       limit,
		Offset:      offset,
	}
}

func getLogsByTeacherParams(teacherID int64, f logFilters, limit, offset int64) queries.GetLogsByCreatedByFilteredParams {
	return queries.GetLogsByCreatedByFilteredParams{
		CreatedBy:   sql.NullInt64{Int64: teacherID, Valid: true},
		Column2:     f.module,
		Module:      f.module,
		Column4:     f.message,
		Column5:     sql.NullString{String: f.message, Valid: true},
		Column6:     f.startDate,
		CreatedAt:   f.startDate,
		Column8:     f.endDate,
		CreatedAt_2: f.endDate,
		Limit:       limit,
		Offset:      offset,
	}
}

func handleSystemLogs(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	sort := parseListSort(r, frontend.ListSortKindSystemLog)
	page := utils.ParsePageQuery(r)
	filters := parseLogFilters(r)
	hideCreatedBy := role == auth.RoleTeacher

	var viewLogs []frontend.SystemLogItem

	if role == auth.RoleTeacher {
		user := auth.GetUser(ctx)
		countParams := countLogsByTeacherParams(user.ID, filters)
		total, err := dbRO.GetQueries().CountLogsByCreatedByFiltered(ctx, countParams)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count logs: %v", err), http.StatusInternalServerError)
			return
		}
		page.Total = total
		allRows, err := dbRO.GetQueries().GetLogsByCreatedByFiltered(ctx, getLogsByTeacherParams(user.ID, filters, total, 0))
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
			return
		}
		sortSystemLogByUserRows(allRows, sort)
		rows := paginateSlice(allRows, page)
		viewLogs = systemLogItemsFromTeacherRows(rows)
	} else {
		total, err := dbRO.GetQueries().CountAllLogsFiltered(ctx, countAllLogsParams(filters))
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count logs: %v", err), http.StatusInternalServerError)
			return
		}
		page.Total = total
		allRows, err := dbRO.GetQueries().GetAllLogsFiltered(ctx, getAllLogsParams(filters, total, 0))
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
			return
		}
		sortSystemLogRows(allRows, sort)
		rows := paginateSlice(allRows, page)
		viewLogs = systemLogItemsFromAllRows(rows)
	}

	params := listQueryParamsWithSort(r, frontend.ListSortKindSystemLog)
	writeHTML(w)
	frontend.SystemLogs(frontend.SystemLogData{
		Logs:           viewLogs,
		HideCreatedBy:  hideCreatedBy,
		Query:          filters.message,
		Module:         filters.module,
		StartDate:      filters.startDate,
		EndDate:        filters.endDate,
		SortBy:         sort.By,
		SortOrder:      string(sort.Order),
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(utils.URL("/logs"), page.Number-1, page.Size, params),
		NextURL:        utils.BuildPageURLAt(utils.URL("/logs"), page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     utils.URL("/logs"),
	}).Render(ctx, w)
}

func countUploadLogsParams(f logFilters) queries.CountUploadLogsFilteredParams {
	return queries.CountUploadLogsFilteredParams{
		Column1:     f.module,
		Module:      f.module,
		Column3:     f.message,
		Column4:     sql.NullString{String: f.message, Valid: true},
		Column5:     f.startDate,
		CreatedAt:   f.startDate,
		Column7:     f.endDate,
		CreatedAt_2: f.endDate,
	}
}

func getUploadLogsParams(f logFilters, limit, offset int64) queries.GetUploadLogsFilteredParams {
	return queries.GetUploadLogsFilteredParams{
		Column1:     f.module,
		Module:      f.module,
		Column3:     f.message,
		Column4:     sql.NullString{String: f.message, Valid: true},
		Column5:     f.startDate,
		CreatedAt:   f.startDate,
		Column7:     f.endDate,
		CreatedAt_2: f.endDate,
		Limit:       limit,
		Offset:      offset,
	}
}

func countUploadLogsByTeacherParams(teacherID int64, f logFilters) queries.CountUploadLogsByCreatedByFilteredParams {
	return queries.CountUploadLogsByCreatedByFilteredParams{
		CreatedBy:   sql.NullInt64{Int64: teacherID, Valid: true},
		Column2:     f.module,
		Module:      f.module,
		Column4:     f.message,
		Column5:     sql.NullString{String: f.message, Valid: true},
		Column6:     f.startDate,
		CreatedAt:   f.startDate,
		Column8:     f.endDate,
		CreatedAt_2: f.endDate,
	}
}

func getUploadLogsByTeacherParams(teacherID int64, f logFilters, limit, offset int64) queries.GetUploadLogsByCreatedByFilteredParams {
	return queries.GetUploadLogsByCreatedByFilteredParams{
		CreatedBy:   sql.NullInt64{Int64: teacherID, Valid: true},
		Column2:     f.module,
		Module:      f.module,
		Column4:     f.message,
		Column5:     sql.NullString{String: f.message, Valid: true},
		Column6:     f.startDate,
		CreatedAt:   f.startDate,
		Column8:     f.endDate,
		CreatedAt_2: f.endDate,
		Limit:       limit,
		Offset:      offset,
	}
}

func uploadLogItemsFromAllRows(rows []queries.GetUploadLogsFilteredRow) []frontend.UploadLogItem {
	viewLogs := make([]frontend.UploadLogItem, len(rows))
	for i, l := range rows {
		viewLogs[i] = mapUploadLogItemFromFiltered(l)
	}
	return viewLogs
}

func uploadLogItemsFromTeacherRows(rows []queries.GetUploadLogsByCreatedByFilteredRow) []frontend.UploadLogItem {
	viewLogs := make([]frontend.UploadLogItem, len(rows))
	for i, l := range rows {
		viewLogs[i] = mapUploadLogItemFromTeacherFiltered(l)
	}
	return viewLogs
}

func handleUploadLogs(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	sort := parseListSort(r, frontend.ListSortKindUploadLog)
	page := utils.ParsePageQuery(r)
	filters := parseLogFilters(r)
	hideTeacher := role == auth.RoleTeacher

	var viewLogs []frontend.UploadLogItem

	if role == auth.RoleTeacher {
		user := auth.GetUser(ctx)
		countParams := countUploadLogsByTeacherParams(user.ID, filters)
		total, err := dbRO.GetQueries().CountUploadLogsByCreatedByFiltered(ctx, countParams)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count upload logs: %v", err), http.StatusInternalServerError)
			return
		}
		page.Total = total
		allRows, err := dbRO.GetQueries().GetUploadLogsByCreatedByFiltered(ctx, getUploadLogsByTeacherParams(user.ID, filters, total, 0))
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to fetch upload logs: %v", err), http.StatusInternalServerError)
			return
		}
		sortUploadLogByUserRows(allRows, sort)
		rows := paginateSlice(allRows, page)
		viewLogs = uploadLogItemsFromTeacherRows(rows)
	} else {
		countParams := countUploadLogsParams(filters)
		total, err := dbRO.GetQueries().CountUploadLogsFiltered(ctx, countParams)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count upload logs: %v", err), http.StatusInternalServerError)
			return
		}
		page.Total = total

		allRows, err := dbRO.GetQueries().GetUploadLogsFiltered(ctx, getUploadLogsParams(filters, total, 0))
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to fetch upload logs: %v", err), http.StatusInternalServerError)
			return
		}
		sortUploadLogRows(allRows, sort)
		rows := paginateSlice(allRows, page)
		viewLogs = uploadLogItemsFromAllRows(rows)
	}

	params := listQueryParamsWithSort(r, frontend.ListSortKindUploadLog)
	writeHTML(w)
	frontend.UploadLogs(frontend.UploadLogData{
		Logs:           viewLogs,
		HideTeacher:    hideTeacher,
		Query:          filters.message,
		Module:         filters.module,
		StartDate:      filters.startDate,
		EndDate:        filters.endDate,
		SortBy:         sort.By,
		SortOrder:      string(sort.Order),
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(utils.URL("/upload-logs"), page.Number-1, page.Size, params),
		NextURL:        utils.BuildPageURLAt(utils.URL("/upload-logs"), page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     utils.URL("/upload-logs"),
	}).Render(ctx, w)
}

func systemLogItemsFromAllRows(rows []queries.GetAllLogsFilteredRow) []frontend.SystemLogItem {
	viewLogs := make([]frontend.SystemLogItem, len(rows))
	for i, l := range rows {
		viewLogs[i] = frontend.SystemLogItem{
			ID:        strconv.FormatInt(l.ID, 10),
			Module:    l.Module,
			Message:   l.Message,
			CreatedBy: l.CreatedByName,
			CreatedAt: l.CreatedAt,
		}
	}
	return viewLogs
}

func systemLogItemsFromTeacherRows(rows []queries.GetLogsByCreatedByFilteredRow) []frontend.SystemLogItem {
	viewLogs := make([]frontend.SystemLogItem, len(rows))
	for i, l := range rows {
		viewLogs[i] = frontend.SystemLogItem{
			ID:        strconv.FormatInt(l.ID, 10),
			Module:    l.Module,
			Message:   l.Message,
			CreatedBy: l.CreatedByName,
			CreatedAt: l.CreatedAt,
		}
	}
	return viewLogs
}

func nullTimeFromDate(s string) sql.NullTime {
	t, err := utils.ParseDatePHT(s)
	if err != nil || t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func processingLogFilterParams(q, startDate, endDate string) queries.CountProcessingLogsFilteredParams {
	return queries.CountProcessingLogsFilteredParams{
		Column1:     q,
		Column2:     sql.NullString{String: q, Valid: true},
		Column3:     startDate,
		CreatedAt:   nullTimeFromDate(startDate),
		Column5:     endDate,
		CreatedAt_2: nullTimeFromDate(endDate),
	}
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	q := r.URL.Query().Get("q")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	sort := parseListSort(r, frontend.ListSortKindProcessingLog)
	page := utils.ParsePageQuery(r)
	filter := processingLogFilterParams(q, startDate, endDate)

	total, err := dbRO.GetQueries().CountProcessingLogsFiltered(ctx, filter)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count logs: %v", err), http.StatusInternalServerError)
		return
	}
	page.Total = total

	allLogs, err := dbRO.GetQueries().GetProcessingLogsFiltered(ctx, queries.GetProcessingLogsFilteredParams{
		Column1:     filter.Column1,
		Column2:     filter.Column2,
		Column3:     filter.Column3,
		CreatedAt:   filter.CreatedAt,
		Column5:     filter.Column5,
		CreatedAt_2: filter.CreatedAt_2,
		Limit:       total,
		Offset:      0,
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
		return
	}
	sortProcessingLogRows(allLogs, sort)
	logs := paginateSlice(allLogs, page)

	viewLogs := make([]frontend.ProcessingLogItem, len(logs))
	for i, l := range logs {
		viewLogs[i] = frontend.ProcessingLogItem{
			ID:             strconv.FormatInt(l.ID, 10),
			GoogleDriveURL: l.GoogleDriveUrl,
			Name:           l.Name,
			Template:       l.Template.String,
			StartDate:      l.StartDate,
			EndDate:        l.EndDate,
			ExcludedRows:   l.ExcludedRows.String,
			UserAgent:      l.Useragent.String,
			OutputPath:     l.OutputPath.String,
			Errors:         l.Errors.String,
			CreatedAt:      utils.FormatNullDateTimeSecondsPHT(l.CreatedAt),
		}
	}

	params := listQueryParamsWithSort(r, frontend.ListSortKindProcessingLog)
	writeHTML(w)
	frontend.ProcessingLogs(frontend.ProcessingLogData{
		Logs:           viewLogs,
		Query:          q,
		StartDate:      startDate,
		EndDate:        endDate,
		SortBy:         sort.By,
		SortOrder:      string(sort.Order),
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(utils.URL("/process-logs"), page.Number-1, page.Size, params),
		NextURL:        utils.BuildPageURLAt(utils.URL("/process-logs"), page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     utils.URL("/process-logs"),
	}).Render(ctx, w)
}
