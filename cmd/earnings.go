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
	"zion-english/internal/utils"
)

func handleMyEarnings(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	teacherID := auth.GetUser(ctx).ID
	rows, emptyMsg, err := loadEarningsCutoffRows(ctx, teacherID)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeHTML(w)
	if err := frontend.MyEarnings(rows, emptyMsg).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTeachersEarnings(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	writeHTML(w)
	if err := frontend.TeachersEarnings().Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTeachersEarningsPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	teacherIDStr := strings.TrimSpace(r.URL.Query().Get("teacherId"))
	if teacherIDStr == "" {
		writeHTML(w)
		if err := frontend.EarningsCutoffTableBody(nil, "Select a teacher to view earnings history.").Render(r.Context(), w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64)
	if err != nil || teacherID <= 0 {
		HttpError(w, ErrInvalidTeacherID.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	_, err = dbRO.GetQueries().GetTeacherByID(ctx, teacherID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			HttpError(w, MsgTeacherNotFound, http.StatusNotFound)
			return
		}
		HttpError(w, "Failed to load teacher", http.StatusInternalServerError)
		return
	}

	rows, emptyMsg, err := loadEarningsCutoffRows(ctx, teacherID)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeHTML(w)
	if err := frontend.EarningsCutoffTableBody(rows, emptyMsg).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadEarningsCutoffRows(ctx context.Context, teacherID int64) ([]frontend.EarningsCutoffRow, string, error) {
	minRaw, err := dbRO.GetQueries().GetMinClassRecordDateForTeacher(ctx, teacherID)
	if err != nil {
		return nil, "", err
	}
	minDate, ok := sqlMinClassRecordDate(minRaw)
	if !ok {
		return nil, "No class records yet.", nil
	}

	rows := buildCutoffEarningsRows(ctx, teacherID, minDate)
	if len(rows) == 0 {
		return nil, "No class records yet.", nil
	}
	return rows, "", nil
}

func sqlMinClassRecordDate(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t), t != ""
	case []byte:
		s := strings.TrimSpace(string(t))
		return s, s != ""
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		return s, s != "" && s != "<nil>"
	}
}

func buildCutoffEarningsRows(ctx context.Context, teacherID int64, minDate string) []frontend.EarningsCutoffRow {
	minT, err := time.ParseInLocation(constants.DateLayout, minDate, constants.LocationPHT)
	if err != nil {
		return nil
	}
	now := time.Now().In(constants.LocationPHT)
	endYear, endMonth := now.Year(), now.Month()
	startYear, startMonth := minT.Year(), minT.Month()

	var rows []frontend.EarningsCutoffRow
	y, m := endYear, endMonth
	for {
		firstPreset, secondPreset := utils.CutoffRangeForMonth(y, m)
		monthLabel := time.Date(y, m, 1, 0, 0, 0, 0, constants.LocationPHT).Format("January 2006")

		rows = append(rows, earningsRowForPreset(ctx, teacherID, monthLabel, frontend.EarningsCutoffKindSecond, secondPreset))
		rows = append(rows, earningsRowForPreset(ctx, teacherID, monthLabel, frontend.EarningsCutoffKindFirst, firstPreset))

		if y == startYear && m == startMonth {
			break
		}
		m--
		if m < time.January {
			m = time.December
			y--
		}
	}
	return rows
}

func earningsRowForPreset(ctx context.Context, teacherID int64, monthLabel string, kind frontend.EarningsCutoffKind, preset string) frontend.EarningsCutoffRow {
	startDate, endDate := utils.CutoffDatesFromPreset(preset)
	row := frontend.EarningsCutoffRow{
		MonthLabel:   monthLabel,
		CutoffKind:   kind,
		PeriodLabel:  formatReportCutoffLabel(startDate, endDate),
		Earnings:     fetchCutoffTotalsByPreset(ctx, preset, teacherID),
		Conducted:    conductedClassCount(ctx, teacherID, startDate, endDate),
	}
	return row
}

func conductedClassCount(ctx context.Context, teacherID int64, startDate, endDate string) int64 {
	statusRows, err := dbRO.GetQueries().CountClassRecordsByStatusAndDateRange(ctx, queries.CountClassRecordsByStatusAndDateRangeParams{
		Date:      startDate,
		Date_2:    endDate,
		Column3:   teacherID,
		TeacherID: teacherID,
	})
	if err != nil {
		return 0
	}
	for _, row := range statusRows {
		if row.Status == string(constants.ClassStatusConducted) {
			return row.Count
		}
	}
	return 0
}
