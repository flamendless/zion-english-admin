package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/classrules"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func seriesIDFromRow(seriesID sql.NullInt64) (int64, bool) {
	if !seriesID.Valid || seriesID.Int64 <= 0 {
		return 0, false
	}
	return seriesID.Int64, true
}

func scheduledClassesForScope(ctx context.Context, existing queries.GetScheduledClassByIDRow, scope constants.SeriesScope) ([]queries.GetScheduledClassesInSeriesFromDateRow, error) {
	seriesID, ok := seriesIDFromRow(existing.SeriesID)
	if !ok {
		return []queries.GetScheduledClassesInSeriesFromDateRow{{
			ID:              existing.ID,
			StudentID:       existing.StudentID,
			TeacherID:       existing.TeacherID,
			ScheduledDate:   existing.ScheduledDate,
			StartTime:       existing.StartTime,
			DurationMinutes: existing.DurationMinutes,
			Rate:            existing.Rate,
			Currency:        existing.Currency,
			IsTrialClass:    existing.IsTrialClass,
			SeriesID:        existing.SeriesID,
			Status:          existing.Status,
		}}, nil
	}
	if scope == constants.SeriesScopeSingle {
		return []queries.GetScheduledClassesInSeriesFromDateRow{{
			ID:              existing.ID,
			StudentID:       existing.StudentID,
			TeacherID:       existing.TeacherID,
			ScheduledDate:   existing.ScheduledDate,
			StartTime:       existing.StartTime,
			DurationMinutes: existing.DurationMinutes,
			Rate:            existing.Rate,
			Currency:        existing.Currency,
			IsTrialClass:    existing.IsTrialClass,
			SeriesID:        existing.SeriesID,
			Status:          existing.Status,
		}}, nil
	}
	rows, err := dbRO.GetQueries().GetScheduledClassesInSeriesFromDate(ctx, queries.GetScheduledClassesInSeriesFromDateParams{
		SeriesID:      sql.NullInt64{Int64: seriesID, Valid: true},
		ScheduledDate: existing.ScheduledDate,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("no scheduled classes found in this series")
	}
	return rows, nil
}

func seriesCounts(ctx context.Context, seriesID sql.NullInt64, fromDate string) (total int64, future int64, err error) {
	id, ok := seriesIDFromRow(seriesID)
	if !ok {
		return 0, 0, nil
	}
	nullSeries := sql.NullInt64{Int64: id, Valid: true}
	total, err = dbRO.GetQueries().CountScheduledClassesBySeries(ctx, nullSeries)
	if err != nil {
		return 0, 0, err
	}
	future, err = dbRO.GetQueries().CountScheduledClassesInSeriesFromDate(ctx, queries.CountScheduledClassesInSeriesFromDateParams{
		SeriesID:      nullSeries,
		ScheduledDate: fromDate,
	})
	if err != nil {
		return 0, 0, err
	}
	return total, future, nil
}

func seriesDates(ctx context.Context, seriesID sql.NullInt64) ([]string, error) {
	id, ok := seriesIDFromRow(seriesID)
	if !ok {
		return nil, nil
	}
	return dbRO.GetQueries().ListScheduledClassDatesBySeries(ctx, sql.NullInt64{Int64: id, Valid: true})
}

func formatSeriesScopeSummary(count int64, action string) string {
	if count <= 1 {
		return ""
	}
	return fmt.Sprintf(" (%d classes %s)", count, action)
}

func enrichScheduledClassItemsSeries(ctx context.Context, items []frontend.ScheduledClassItemData) ([]frontend.ScheduledClassItemData, error) {
	q := dbRO.GetQueries()
	for i := range items {
		if items[i].SeriesID <= 0 {
			continue
		}
		nullSeries := sql.NullInt64{Int64: items[i].SeriesID, Valid: true}
		total, err := q.CountScheduledClassesBySeries(ctx, nullSeries)
		if err != nil {
			return nil, err
		}
		future, err := q.CountScheduledClassesInSeriesFromDate(ctx, queries.CountScheduledClassesInSeriesFromDateParams{
			SeriesID:      nullSeries,
			ScheduledDate: items[i].ScheduledDate,
		})
		if err != nil {
			return nil, err
		}
		items[i].SeriesTotalCount = total
		items[i].SeriesFutureCount = future
	}
	return items, nil
}

func scheduledClassSeriesItemsFromRows(ctx context.Context, rows []queries.ListActiveScheduledClassSeriesRow) ([]frontend.ScheduledClassSeriesItem, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	teacherIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		teacherIDs = append(teacherIDs, row.TeacherID)
	}
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, err
	}

	items := make([]frontend.ScheduledClassSeriesItem, 0, len(rows))
	for _, row := range rows {
		startTime := ""
		if row.StartTime.Valid {
			startTime = row.StartTime.String
		}
		endTime := utils.EndTimeFromStartAndDuration(startTime, row.DurationMinutes)
		avatar := avatarWithTeacherRoles(
			buildTeacherListAvatarProps(
				row.TeacherID,
				row.TeacherFirstName,
				row.TeacherMiddleName,
				row.TeacherLastName,
				row.TeacherAssignedColor,
				row.TeacherProfilePicture,
			),
			rolesMap[row.TeacherID],
		)
		nullSeries := sql.NullInt64{Int64: row.SeriesID, Valid: true}
		_, future, err := seriesCounts(ctx, nullSeries, row.AnchorDate)
		if err != nil {
			return nil, err
		}
		dates, err := seriesDates(ctx, nullSeries)
		if err != nil {
			return nil, err
		}
		item := frontend.ScheduledClassItemData{
			ID:              row.AnchorClassID,
			StudentID:       row.StudentID,
			TeacherID:       row.TeacherID,
			StudentName:     row.StudentName,
			TeacherName:     row.TeacherName,
			TeacherAvatar:   avatar,
			ScheduledDate:   row.AnchorDate,
			StartTime:       startTime,
			EndTime:         endTime,
			DurationMinutes: row.DurationMinutes,
			Rate:            row.Rate,
			Currency:        row.Currency,
			Status:          constants.ScheduledClassStatusScheduled,
			TimeRange:       frontend.FormatScheduledClassTimeRange(startTime, endTime, row.DurationMinutes),
			DeleteFrom:        frontend.ClassActionContextScheduleSeries,
			SeriesID:        row.SeriesID,
			SeriesTotalCount: row.ScheduledCount,
			SeriesFutureCount: future,
		}
		items = append(items, frontend.ScheduledClassSeriesItem{
			Item:           item,
			Dates:          dates,
			ScheduledCount: row.ScheduledCount,
		})
	}
	return items, nil
}

func loadScheduledClassSeriesItems(ctx context.Context, role auth.Role, userID int64) ([]frontend.ScheduledClassSeriesItem, bool, error) {
	showTeacher := auth.HasAdminAccess(role)
	teacherFilter := int64(0)
	if auth.IsTeacherScoped(role) {
		teacherFilter = userID
	}
	rows, err := dbRO.GetQueries().ListActiveScheduledClassSeries(ctx, queries.ListActiveScheduledClassSeriesParams{
		Column1:   teacherFilter,
		TeacherID: teacherFilter,
	})
	if err != nil {
		return nil, showTeacher, err
	}
	items, err := scheduledClassSeriesItemsFromRows(ctx, rows)
	if err != nil {
		return nil, showTeacher, err
	}
	return items, showTeacher, nil
}

func parseEditSeriesDateList(rawDates []string, allowedExisting []string) ([]string, error) {
	allowedPast := make(map[string]struct{}, len(allowedExisting))
	for _, d := range allowedExisting {
		allowedPast[d] = struct{}{}
	}
	seen := make(map[string]struct{}, len(rawDates))
	dates := make([]string, 0, len(rawDates))
	today := utils.TodayPHT()
	for _, raw := range rawDates {
		date := strings.TrimSpace(raw)
		if date == "" {
			continue
		}
		if _, err := utils.ParseDatePHT(date); err != nil {
			return nil, errors.New("invalid date format")
		}
		if date < today {
			if _, ok := allowedPast[date]; !ok {
				return nil, fmt.Errorf("date %s cannot be in the past", date)
			}
		}
		if _, ok := seen[date]; ok {
			continue
		}
		seen[date] = struct{}{}
		dates = append(dates, date)
	}
	if len(dates) == 0 {
		return nil, errors.New("select at least one date for this series")
	}
	if len(dates) > constants.MaxRepeatScheduleDates {
		return nil, fmt.Errorf("select at most %d dates", constants.MaxRepeatScheduleDates)
	}
	sort.Strings(dates)
	return dates, nil
}

func parseEditSeriesDates(r *http.Request, allowedExisting []string) ([]string, error) {
	rawDates := r.Form["scheduled_dates"]
	if len(rawDates) == 0 {
		return nil, nil
	}
	return parseEditSeriesDateList(rawDates, allowedExisting)
}

func parseEditSeriesConfirmDates(r *http.Request, allowedExisting []string) ([]string, error) {
	if dates := r.Form["confirm_dates"]; len(dates) > 0 {
		return parseEditSeriesDateList(dates, allowedExisting)
	}
	return parseEditSeriesDates(r, allowedExisting)
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa := append([]string(nil), a...)
	bb := append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}

func diffStringSlices(old, new []string) (removed, added []string) {
	oldSet := make(map[string]struct{}, len(old))
	newSet := make(map[string]struct{}, len(new))
	for _, d := range old {
		oldSet[d] = struct{}{}
	}
	for _, d := range new {
		newSet[d] = struct{}{}
	}
	for d := range oldSet {
		if _, ok := newSet[d]; !ok {
			removed = append(removed, d)
		}
	}
	for d := range newSet {
		if _, ok := oldSet[d]; !ok {
			added = append(added, d)
		}
	}
	sort.Strings(removed)
	sort.Strings(added)
	return removed, added
}

func filterDatesFromOn(dates []string, fromDate string) []string {
	out := make([]string, 0, len(dates))
	for _, d := range dates {
		if d >= fromDate {
			out = append(out, d)
		}
	}
	sort.Strings(out)
	return out
}

func targetStartTime(target queries.GetScheduledClassesInSeriesFromDateRow) string {
	if target.StartTime.Valid {
		return target.StartTime.String
	}
	return ""
}

func scheduledClassesInSeries(ctx context.Context, seriesID sql.NullInt64) ([]queries.GetScheduledClassesInSeriesFromDateRow, error) {
	id, ok := seriesIDFromRow(seriesID)
	if !ok {
		return nil, nil
	}
	rows, err := dbRO.GetQueries().GetScheduledClassesInSeriesFromDate(ctx, queries.GetScheduledClassesInSeriesFromDateParams{
		SeriesID:      sql.NullInt64{Int64: id, Valid: true},
		ScheduledDate: "1970-01-01",
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ScheduledDate == rows[j].ScheduledDate {
			return rows[i].ID < rows[j].ID
		}
		return rows[i].ScheduledDate < rows[j].ScheduledDate
	})
	return rows, nil
}

func seriesDateTargets(scope constants.SeriesScope, anchorDate string, classes []queries.GetScheduledClassesInSeriesFromDateRow) []queries.GetScheduledClassesInSeriesFromDateRow {
	if scope == constants.SeriesScopeFuture {
		out := make([]queries.GetScheduledClassesInSeriesFromDateRow, 0, len(classes))
		for _, class := range classes {
			if class.ScheduledDate >= anchorDate {
				out = append(out, class)
			}
		}
		return out
	}
	return classes
}

func dateInList(date string, dates []string) bool {
	for _, d := range dates {
		if d == date {
			return true
		}
	}
	return false
}

func finalizeSeriesConfirmDates(scope constants.SeriesScope, anchorDate string, allSeriesDates, confirmDates []string) ([]string, error) {
	if scope == constants.SeriesScopeSingle {
		return confirmDates, nil
	}
	for _, date := range allSeriesDates {
		if date < anchorDate && !dateInList(date, confirmDates) {
			return nil, errors.New("past sessions in this series cannot be removed when applying to this and future classes only")
		}
	}
	merged := make(map[string]struct{}, len(confirmDates)+len(allSeriesDates))
	for _, date := range confirmDates {
		merged[date] = struct{}{}
	}
	for _, date := range allSeriesDates {
		if date < anchorDate {
			merged[date] = struct{}{}
		}
	}
	out := make([]string, 0, len(merged))
	for date := range merged {
		out = append(out, date)
	}
	sort.Strings(out)
	return out, nil
}

func seriesFieldTargetIDs(scope constants.SeriesScope, existing queries.GetScheduledClassByIDRow, scopeTargets []queries.GetScheduledClassesInSeriesFromDateRow) map[int64]bool {
	targets := make(map[int64]bool)
	if scope == constants.SeriesScopeSingle {
		targets[existing.ID] = true
		return targets
	}
	for _, target := range scopeTargets {
		targets[target.ID] = true
	}
	return targets
}

func buildSeriesEditInput(target queries.GetScheduledClassesInSeriesFromDateRow, newDate string, editInput scheduledClassEditInput, fieldTargets map[int64]bool) scheduledClassEditInput {
	if fieldTargets[target.ID] {
		out := editInput
		out.ScheduledDate = newDate
		return out
	}
	return scheduledClassEditInput{
		StudentID:       target.StudentID,
		ScheduledDate:   newDate,
		StartTime:       targetStartTime(target),
		DurationMinutes: target.DurationMinutes,
		Rate:            target.Rate,
		Currency:        target.Currency,
		IsTrialClass:    target.IsTrialClass != 0,
	}
}

func learningMaterialIDsForTarget(targetID int64, fieldTargets map[int64]bool, lmIDs []int64) []int64 {
	if fieldTargets[targetID] {
		return lmIDs
	}
	return nil
}

func seriesEditFieldsChanged(target queries.GetScheduledClassesInSeriesFromDateRow, input scheduledClassEditInput) bool {
	start := targetStartTime(target)
	return input.StudentID != target.StudentID ||
		input.StartTime != start ||
		input.DurationMinutes != target.DurationMinutes ||
		input.Rate != target.Rate ||
		input.Currency != target.Currency ||
		input.IsTrialClass != (target.IsTrialClass != 0)
}

func validateSeriesEditRow(ctx context.Context, user auth.User, classID int64, teacherID int64, input scheduledClassEditInput) (bool, string) {
	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
		ScheduleID:      classID,
		StudentID:       input.StudentID,
		TeacherID:       teacherID,
		Date:            input.ScheduledDate,
		StartTime:       input.StartTime,
		DurationMinutes: input.DurationMinutes,
	})
	if err != nil {
		return false, err.Error()
	}
	return true, "Ready"
}

func previewSeriesEditRows(ctx context.Context, user auth.User, existing queries.GetScheduledClassByIDRow, scope constants.SeriesScope, selectedDates []string, editInput scheduledClassEditInput) ([]frontend.ScheduleSeriesEditPreviewRow, error) {
	classes, err := scheduledClassesInSeries(ctx, existing.SeriesID)
	if err != nil {
		return nil, err
	}
	classByDate := make(map[string]queries.GetScheduledClassesInSeriesFromDateRow, len(classes))
	for _, class := range classes {
		classByDate[class.ScheduledDate] = class
	}

	if scope == constants.SeriesScopeFuture {
		for _, class := range classes {
			if class.ScheduledDate < existing.ScheduledDate && !dateInList(class.ScheduledDate, selectedDates) {
				return nil, errors.New("past sessions in this series cannot be removed when applying to this and future classes only")
			}
		}
	}

	scopeTargets, err := scheduledClassesForScope(ctx, existing, scope)
	if err != nil {
		return nil, err
	}
	fieldTargets := seriesFieldTargetIDs(scope, existing, scopeTargets)
	dateTargets := seriesDateTargets(scope, existing.ScheduledDate, classes)
	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}

	rows := make([]frontend.ScheduleSeriesEditPreviewRow, 0, len(selectedDates)+len(dateTargets))
	for _, target := range dateTargets {
		if dateInList(target.ScheduledDate, selectedDates) {
			continue
		}
		rows = append(rows, frontend.ScheduleSeriesEditPreviewRow{
			Date:    target.ScheduledDate,
			Action:  "Remove",
			Ready:   true,
			Message: "Will remove this session",
		})
	}

	for _, date := range selectedDates {
		if scope == constants.SeriesScopeFuture && date < existing.ScheduledDate {
			if _, ok := classByDate[date]; ok {
				continue
			}
			rows = append(rows, frontend.ScheduleSeriesEditPreviewRow{
				Date:    date,
				Action:  "Add",
				Ready:   false,
				Message: "Cannot add past dates when applying to this and future classes only",
			})
			continue
		}

		if class, ok := classByDate[date]; ok {
			input := buildSeriesEditInput(class, date, editInput, fieldTargets)
			action := "Keep"
			if fieldTargets[class.ID] && seriesEditFieldsChanged(class, input) {
				action = "Update"
			}
			ready, message := validateSeriesEditRow(ctx, user, class.ID, existing.TeacherID, input)
			rows = append(rows, frontend.ScheduleSeriesEditPreviewRow{
				Date:    date,
				Action:  action,
				Ready:   ready,
				Message: message,
			})
			continue
		}

		row := frontend.ScheduleSeriesEditPreviewRow{Date: date, Action: "Add"}
		if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
			StudentID:       editInput.StudentID,
			TeacherID:       existing.TeacherID,
			Date:            date,
			StartTime:       editInput.StartTime,
			DurationMinutes: editInput.DurationMinutes,
		}); err != nil {
			row.Message = err.Error()
		} else {
			row.Ready = true
			row.Message = "Ready"
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date < rows[j].Date
	})
	return rows, nil
}

func applySeriesScheduleEdit(ctx context.Context, user auth.User, existing queries.GetScheduledClassByIDRow, scope constants.SeriesScope, confirmDates, allSeriesDates []string, editInput scheduledClassEditInput, lmIDs []int64, existingStart string) (int, error) {
	finalDates, err := finalizeSeriesConfirmDates(scope, existing.ScheduledDate, allSeriesDates, confirmDates)
	if err != nil {
		return 0, err
	}

	classes, err := scheduledClassesInSeries(ctx, existing.SeriesID)
	if err != nil {
		return 0, err
	}
	scopeTargets, err := scheduledClassesForScope(ctx, existing, scope)
	if err != nil {
		return 0, err
	}
	fieldTargets := seriesFieldTargetIDs(scope, existing, scopeTargets)
	dateTargets := seriesDateTargets(scope, existing.ScheduledDate, classes)

	scopeConfirmDates := finalDates
	if scope == constants.SeriesScopeFuture {
		scopeConfirmDates = filterDatesFromOn(finalDates, existing.ScheduledDate)
	}

	confirmSet := make(map[string]struct{}, len(scopeConfirmDates))
	for _, date := range scopeConfirmDates {
		confirmSet[date] = struct{}{}
	}

	classByDate := make(map[string]queries.GetScheduledClassesInSeriesFromDateRow, len(classes))
	for _, class := range classes {
		classByDate[class.ScheduledDate] = class
	}

	affected := 0
	for _, target := range dateTargets {
		if _, keep := confirmSet[target.ScheduledDate]; keep {
			continue
		}
		if err := deleteScheduledClassByID(ctx, user, target.ID, "Removed from series during edit"); err != nil {
			return affected, err
		}
		delete(classByDate, target.ScheduledDate)
		affected++
	}

	for _, date := range scopeConfirmDates {
		if class, ok := classByDate[date]; ok {
			input := buildSeriesEditInput(class, date, editInput, fieldTargets)
			if fieldTargets[class.ID] {
				if class.ID == existing.ID {
					if err := validateScheduleDateTimeChange(existing.ScheduledDate, existingStart, date, input.StartTime); err != nil {
						return affected, err
					}
				} else if seriesEditFieldsChanged(class, input) {
					if err := validateScheduleDateTimeChange(class.ScheduledDate, targetStartTime(class), date, input.StartTime); err != nil {
						return affected, err
					}
				}
			}
			lm := learningMaterialIDsForTarget(class.ID, fieldTargets, lmIDs)
			if err := applyScheduledClassEdit(ctx, user, class, input, lm); err != nil {
				return affected, err
			}
			affected++
			continue
		}

		params := scheduledClassCreateParams{
			StudentID:       editInput.StudentID,
			TeacherID:       existing.TeacherID,
			ScheduledDate:   date,
			StartTime:       editInput.StartTime,
			DurationMinutes: editInput.DurationMinutes,
			Rate:            editInput.Rate,
			Currency:        editInput.Currency,
			IsTrialClass:    editInput.IsTrialClass,
			SeriesID:        existing.SeriesID,
		}
		scheduleID, err := insertScheduledClassRow(ctx, dbRW.GetQueries(), user, params)
		if err != nil {
			return affected, err
		}
		if err := runScheduledClassSideEffects(ctx, user, scheduleID, params, lmIDs, "series edit create"); err != nil {
			return affected, err
		}
		affected++
	}

	return affected, nil
}

func handleScheduleSeries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)

	series, showTeacher, err := loadScheduledClassSeriesItems(ctx, role, user.ID)
	if err != nil {
		logs.Log().Error("load repeating scheduled class series", zap.Error(err))
		HttpError(w, "Failed to load repeating scheduled classes", http.StatusInternalServerError)
		return
	}

	if err := frontend.ScheduleSeries(frontend.ScheduleSeriesData{
		Series:            series,
		ShowSeriesTeacher: showTeacher,
	}).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}
