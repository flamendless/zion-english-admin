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
	"zion-english/internal/sqlconsole"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func handleSQLConsole(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	saved, err := dbRO.GetQueries().ListSQLConsoleSavedQueries(ctx)
	if err != nil {
		HttpError(w, "Failed to load page", http.StatusInternalServerError)
		logs.Log().Error("list sql console saved queries", zap.Error(err))
		return
	}

	page := utils.Page{Number: 1, Size: constants.SQLConsoleHistoryPageSize}
	history, historyPage, err := fetchSQLConsoleHistory(ctx, page)
	if err != nil {
		HttpError(w, "Failed to load page", http.StatusInternalServerError)
		logs.Log().Error("list sql console history", zap.Error(err))
		return
	}

	data := frontend.SQLConsoleData{
		Saved:   mapSQLConsoleSavedItems(saved),
		History: history,
		HistoryPage: historyPage,
		HistoryURL: utils.URL("/sql-console/history"),
	}
	if err := frontend.SQLConsole(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSQLConsoleRun(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		renderSQLConsoleResultsError(w, r, ErrSQLConsoleInvalidForm.Error())
		return
	}

	rawQuery := r.FormValue("query")
	savedID := parseSQLConsoleSavedQueryID(r.FormValue("saved_query_id"))
	user := auth.GetUser(ctx)
	start := time.Now()

	normalized, valErr := sqlconsole.ValidateReadOnlyQuery(rawQuery)
	if valErr != nil {
		stats := sqlconsole.Stats{
			DurationMs:  time.Since(start).Milliseconds(),
			QueryLength: len(strings.TrimSpace(rawQuery)),
			RowLimit:    constants.SQLConsoleDefaultRowLimit,
		}
		logText := normalized
		if logText == "" {
			logText = strings.TrimSpace(rawQuery)
		}
		insertSQLConsoleQueryLog(ctx, user, logText, savedID, constants.SQLConsoleOutcomeRejected, valErr.Error(), stats)
		renderSQLConsoleResultsError(w, r, valErr.Error())
		return
	}

	result, execErr := sqlconsole.Execute(ctx, dbRO.GetDB(), normalized, sqlconsole.DefaultOptions())
	if execErr != nil {
		msg := execErr.Error()
		insertSQLConsoleQueryLog(ctx, user, normalized, savedID, constants.SQLConsoleOutcomeFailed, msg, result.Stats)
		renderSQLConsoleResultsError(w, r, msg)
		return
	}

	insertSQLConsoleQueryLog(ctx, user, normalized, savedID, constants.SQLConsoleOutcomeSucceeded, "", result.Stats)
	if err := frontend.SQLConsoleResults(mapSQLConsoleResult(result, "")).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSQLConsoleSaved(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		HttpError(w, ErrSQLConsoleInvalidForm.Error(), http.StatusBadRequest)
		return
	}

	user := auth.GetUser(ctx)
	action := strings.TrimSpace(r.FormValue("action"))
	switch action {
	case "create", "update":
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			HttpError(w, sqlconsole.ErrSQLSavedNameRequired.Error(), http.StatusBadRequest)
			return
		}
		if len(name) > constants.SQLConsoleMaxSavedNameLen {
			HttpError(w, sqlconsole.ErrSQLSavedNameTooLong.Error(), http.StatusBadRequest)
			return
		}
		queryText, err := sqlconsole.ValidateReadOnlyQuery(r.FormValue("query"))
		if err != nil {
			HttpError(w, err.Error(), http.StatusBadRequest)
			return
		}
		createdBy, createdByName := sqlConsoleActorFields(user)
		switch action {
		case "create":
			id, err := dbRW.GetQueries().InsertSQLConsoleSavedQuery(ctx, queries.InsertSQLConsoleSavedQueryParams{
				Name:           name,
				QueryText:      queryText,
				CreatedBy:      createdBy,
				CreatedByName:  createdByName,
			})
			if err != nil {
				if isSQLConsoleUniqueViolation(err) {
					HttpError(w, sqlconsole.ErrSQLSavedNameDuplicate.Error(), http.StatusBadRequest)
					return
				}
				HttpError(w, "Failed to save", http.StatusInternalServerError)
				return
			}
			insertAuditLogAs(ctx, user, "sql-console", fmt.Sprintf("created saved query #%d (%s)", id, name))
		case "update":
			id, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("id")), 10, 64)
			if err != nil || id <= 0 {
				HttpError(w, sqlconsole.ErrSQLSavedQueryNotFound.Error(), http.StatusBadRequest)
				return
			}
			if _, err := dbRO.GetQueries().GetSQLConsoleSavedQueryByID(ctx, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					HttpError(w, sqlconsole.ErrSQLSavedQueryNotFound.Error(), http.StatusNotFound)
					return
				}
				HttpError(w, "Failed to save", http.StatusInternalServerError)
				return
			}
			if err := dbRW.GetQueries().UpdateSQLConsoleSavedQuery(ctx, queries.UpdateSQLConsoleSavedQueryParams{
				Name:      name,
				QueryText: queryText,
				ID:        id,
			}); err != nil {
				if isSQLConsoleUniqueViolation(err) {
					HttpError(w, sqlconsole.ErrSQLSavedNameDuplicate.Error(), http.StatusBadRequest)
					return
				}
				HttpError(w, "Failed to save", http.StatusInternalServerError)
				return
			}
			insertAuditLogAs(ctx, user, "sql-console", fmt.Sprintf("updated saved query #%d (%s)", id, name))
		}
	case "delete":
		id, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("id")), 10, 64)
		if err != nil || id <= 0 {
			HttpError(w, sqlconsole.ErrSQLSavedQueryNotFound.Error(), http.StatusBadRequest)
			return
		}
		row, err := dbRO.GetQueries().GetSQLConsoleSavedQueryByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				HttpError(w, sqlconsole.ErrSQLSavedQueryNotFound.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, "Failed to delete", http.StatusInternalServerError)
			return
		}
		if err := dbRW.GetQueries().DeleteSQLConsoleSavedQuery(ctx, id); err != nil {
			HttpError(w, "Failed to delete", http.StatusInternalServerError)
			return
		}
		insertAuditLogAs(ctx, user, "sql-console", fmt.Sprintf("deleted saved query #%d (%s)", id, row.Name))
	default:
		HttpError(w, sqlconsole.ErrSQLSavedInvalidAction.Error(), http.StatusBadRequest)
		return
	}

	setHXRedirect(w, "/sql-console")
}

func handleSQLConsoleHistoryPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	page := utils.ParsePageQuery(r)
	if page.Size <= 0 || page.Size > utils.MaxPageSize {
		page.Size = constants.SQLConsoleHistoryPageSize
	}

	items, pageData, err := fetchSQLConsoleHistory(ctx, page)
	if err != nil {
		HttpError(w, "Failed to load page", http.StatusInternalServerError)
		return
	}

	writeHTML(w)
	if err := frontend.SQLConsoleHistoryPanel(items, pageData, utils.URL("/sql-console/history")).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func fetchSQLConsoleHistory(ctx context.Context, page utils.Page) ([]frontend.SQLConsoleHistoryItem, frontend.PaginationData, error) {
	total, err := dbRO.GetQueries().CountSQLConsoleQueryLog(ctx)
	if err != nil {
		return nil, frontend.PaginationData{}, err
	}
	page.Total = total
	rows, err := dbRO.GetQueries().ListSQLConsoleQueryLog(ctx, queries.ListSQLConsoleQueryLogParams{
		Limit:  int64(page.Size),
		Offset: int64(page.Offset()),
	})
	if err != nil {
		return nil, frontend.PaginationData{}, err
	}
	return mapSQLConsoleHistoryItems(rows), frontend.BuildPaginationData(page.Number, page.Size, total), nil
}

func insertSQLConsoleQueryLog(ctx context.Context, user auth.User, queryText string, savedID sql.NullInt64, outcome constants.SQLConsoleQueryOutcome, errMsg string, stats sqlconsole.Stats) {
	executedBy, executedByName := sqlConsoleActorFields(user)
	var errNull sql.NullString
	if errMsg != "" {
		errNull = sql.NullString{String: errMsg, Valid: true}
	}
	truncated := int64(0)
	if stats.Truncated {
		truncated = 1
	}
	_, err := dbRW.GetQueries().InsertSQLConsoleQueryLog(ctx, queries.InsertSQLConsoleQueryLogParams{
		QueryText:      queryText,
		SavedQueryID:   savedID,
		Outcome:        string(outcome),
		ErrorMessage:   errNull,
		DurationMs:     sqlNullInt64(stats.DurationMs, stats.DurationMs >= 0),
		ColumnCount:    sqlNullInt64(int64(stats.ColumnCount), outcome == constants.SQLConsoleOutcomeSucceeded),
		RowsReturned:   sqlNullInt64(int64(stats.RowsReturned), outcome == constants.SQLConsoleOutcomeSucceeded),
		QueryLength:    sqlNullInt64(int64(stats.QueryLength), stats.QueryLength > 0),
		Truncated:      truncated,
		ExecutedBy:     executedBy,
		ExecutedByName: executedByName,
	})
	if err != nil {
		logs.Log().Error("insert sql console query log", zap.Error(err))
	}
}

func sqlConsoleActorFields(actor auth.User) (sql.NullInt64, string) {
	var by sql.NullInt64
	if actor.ID > 0 {
		by = sql.NullInt64{Int64: actor.ID, Valid: true}
	}
	name := strings.TrimSpace(actor.Name)
	if name == "" {
		name = "unknown"
	}
	return by, name
}

func sqlNullInt64(v int64, ok bool) sql.NullInt64 {
	if !ok {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v, Valid: true}
}

func parseSQLConsoleSavedQueryID(raw string) sql.NullInt64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sql.NullInt64{}
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: id, Valid: true}
}

func mapSQLConsoleSavedItems(rows []queries.TblSqlConsoleSavedQuery) []frontend.SQLConsoleSavedItem {
	out := make([]frontend.SQLConsoleSavedItem, len(rows))
	for i, row := range rows {
		out[i] = frontend.SQLConsoleSavedItem{
			ID:        strconv.FormatInt(row.ID, 10),
			Name:      row.Name,
			QueryText: row.QueryText,
		}
	}
	return out
}

func mapSQLConsoleHistoryItems(rows []queries.TblSqlConsoleQueryLog) []frontend.SQLConsoleHistoryItem {
	out := make([]frontend.SQLConsoleHistoryItem, len(rows))
	for i, row := range rows {
		stats := sqlConsoleHistoryStatsLine(row)
		out[i] = frontend.SQLConsoleHistoryItem{
			ID:           strconv.FormatInt(row.ID, 10),
			QueryPreview: sqlConsoleQueryPreview(row.QueryText),
			QueryText:    row.QueryText,
			Outcome:      constants.SQLConsoleQueryOutcome(row.Outcome),
			StatsLine:    stats,
			CreatedAt:    utils.FormatPlanDateTimePHT(row.CreatedAt),
		}
	}
	return out
}

func sqlConsoleQueryPreview(q string) string {
	q = strings.TrimSpace(q)
	if len(q) <= constants.SQLConsoleHistoryPreviewLen {
		return q
	}
	return q[:constants.SQLConsoleHistoryPreviewLen] + "..."
}

func sqlConsoleHistoryStatsLine(row queries.TblSqlConsoleQueryLog) string {
	if row.Outcome != string(constants.SQLConsoleOutcomeSucceeded) {
		if row.DurationMs.Valid {
			return fmt.Sprintf("%d ms", row.DurationMs.Int64)
		}
		return ""
	}
	parts := []string{}
	if row.DurationMs.Valid {
		parts = append(parts, fmt.Sprintf("%d ms", row.DurationMs.Int64))
	}
	if row.RowsReturned.Valid {
		parts = append(parts, fmt.Sprintf("%d rows", row.RowsReturned.Int64))
	}
	if row.Truncated == 1 {
		parts = append(parts, "truncated")
	}
	return strings.Join(parts, ", ")
}

func mapSQLConsoleResult(result sqlconsole.Result, errMsg string) frontend.SQLConsoleResultView {
	return frontend.SQLConsoleResultView{
		Columns: result.Columns,
		Rows:    result.Rows,
		Stats: frontend.SQLConsoleStatsView{
			DurationMs:   result.Stats.DurationMs,
			ColumnCount:  result.Stats.ColumnCount,
			RowsReturned: result.Stats.RowsReturned,
			RowLimit:     result.Stats.RowLimit,
			Truncated:    result.Stats.Truncated,
			QueryLength:  result.Stats.QueryLength,
		},
		ErrorMessage: errMsg,
	}
}

func renderSQLConsoleResultsError(w http.ResponseWriter, r *http.Request, message string) {
	writeHTML(w)
	if err := frontend.SQLConsoleResults(mapSQLConsoleResult(sqlconsole.Result{}, message)).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func isSQLConsoleUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "constraint")
}
