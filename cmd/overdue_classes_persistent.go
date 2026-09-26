package cmd

import (
	"context"
	"net/http"

	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
)

const persistentOverdueClassesListLimit = frontend.PersistentOverdueClassesPanelCap

func loadTeacherOverdueClassesThisWeek(ctx context.Context, teacherID int64) (int64, []queries.ListOverdueScheduledClassesByDateRangeRow, error) {
	weekStart, weekEnd := weekRange()
	cutoff := overdueCutoffPHT(ctx)
	q := dbRO.GetQueries()

	total, err := q.CountOverdueScheduledClassesByDateRange(ctx, queries.CountOverdueScheduledClassesByDateRangeParams{
		Column1:         teacherID,
		TeacherID:       teacherID,
		ScheduledDate:   weekStart,
		ScheduledDate_2: weekEnd,
		Datetime:        cutoff,
	})
	if err != nil {
		return 0, nil, err
	}
	if total == 0 {
		return 0, nil, nil
	}

	rows, err := q.ListOverdueScheduledClassesByDateRange(ctx, queries.ListOverdueScheduledClassesByDateRangeParams{
		Column1:         teacherID,
		TeacherID:       teacherID,
		ScheduledDate:   weekStart,
		ScheduledDate_2: weekEnd,
		Datetime:        cutoff,
		Limit:           persistentOverdueClassesListLimit,
	})
	if err != nil {
		return 0, nil, err
	}
	return total, rows, nil
}

func handlePersistentOverdueClassesPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	if !auth.IsTeacherScoped(role) {
		return
	}

	user := auth.GetUser(ctx)
	total, rows, err := loadTeacherOverdueClassesThisWeek(ctx, user.ID)
	if err != nil || total == 0 {
		return
	}

	weekStart, weekEnd := weekRange()
	panelRows := make([]frontend.OverdueClassPanelItem, 0, len(rows))
	for _, row := range rows {
		startTime := ""
		if row.StartTime.Valid {
			startTime = row.StartTime.String
		}
		panelRows = append(panelRows, frontend.OverdueClassPanelItem{
			ID:              row.ID,
			StudentName:     row.StudentName,
			ScheduledDate:   row.ScheduledDate,
			StartTime:       startTime,
			DurationMinutes: row.DurationMinutes,
		})
	}
	panel := frontend.BuildOverdueClassesPersistentPanel(total, panelRows, weekStart, weekEnd)
	writeHTML(w)
	if err := frontend.PersistentPanel(panel).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}
