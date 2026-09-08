package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/classrules"
	"zion-english/internal/database/queries"
	"zion-english/internal/notifications"
	"zion-english/internal/utils"
)

func handleScheduleRepeat(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleScheduleRepeatPage(w, r)
	case http.MethodPost:
		handleScheduleRepeatConfirm(w, r)
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleScheduleRepeatPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	base, err := parseScheduledClassBase(r, user, role)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	dates, err := parseScheduledDates(r)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rows := previewRepeatScheduleRows(ctx, user, base, dates)
	previewRows := make([]frontend.ScheduleRepeatPreviewRow, 0, len(rows))
	for _, row := range rows {
		previewRows = append(previewRows, frontend.ScheduleRepeatPreviewRow{
			Date:    row.Date,
			Ready:   row.Ready,
			Message: row.Message,
		})
	}

	data := frontend.ScheduleRepeatPreviewData{
		Base: frontend.ScheduleRepeatFormSnapshot{
			TeacherID:   fmt.Sprintf("%d", base.TeacherID),
			StudentID:   fmt.Sprintf("%d", base.StudentID),
			StartTime:   base.StartTime,
			EndTime:     utils.EndTimeFromStartAndDuration(base.StartTime, base.DurationMinutes),
			Rate:        base.Rate,
			Currency:    base.Currency,
			IsTrialClass: base.IsTrialClass,
			LockTeacher: auth.IsTeacherScoped(role),
		},
		Rows:              previewRows,
		LearningMaterialIDs: parseLearningMaterialIDs(r),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ScheduleRepeatPreview(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleScheduleRepeatPage(w http.ResponseWriter, r *http.Request) {
	data := frontend.ScheduleRepeatData{
		TodayPHT: utils.TodayPHT(),
	}
	role := auth.GetRole(r.Context())
	data.IsSuperuser = auth.HasAdminAccess(role)
	if auth.IsTeacherScoped(role) {
		user := auth.GetUser(r.Context())
		data.LockTeacher = true
		data.TeacherID = fmt.Sprintf("%d", user.ID)
		data.TeacherName = user.Name
	}
	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ScheduleRepeat(data).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleScheduleRepeatConfirm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	base, err := parseScheduledClassBase(r, user, role)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	dates, err := parseConfirmedScheduledDates(r)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	for _, date := range dates {
		if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
			StudentID:       base.StudentID,
			TeacherID:       base.TeacherID,
			Date:            date,
			StartTime:       base.StartTime,
			DurationMinutes: base.DurationMinutes,
		}); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	lmIDs := parseLearningMaterialIDs(r)
	createdIDs, err := createRepeatScheduledClasses(ctx, user, base, dates)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	for i, scheduleID := range createdIDs {
		params := base
		params.ScheduledDate = dates[i]
		if err := runScheduledClassSideEffects(ctx, user, scheduleID, params, lmIDs, "schedule repeat create"); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	insertAuditLogAs(ctx, user, "schedule", fmt.Sprintf("created repeating series with %d classes for student id %d (teacher id %d)", len(createdIDs), base.StudentID, base.TeacherID))
	notifyCrossParty(ctx, user, base.TeacherID, teacherNameByID(ctx, base.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Created %d scheduled classes", len(createdIDs)))

	dateLabels := make([]string, 0, len(dates))
	for _, date := range dates {
		dateLabels = append(dateLabels, frontend.FormatScheduledClassDateDisplay(date))
	}
	message := fmt.Sprintf("Created %d classes across %s", len(createdIDs), strings.Join(dateLabels, ", "))
	setSuccessFlash(w, message)
	w.Header().Set("HX-Redirect", utils.URL("/schedule/repeat"))
}

func createRepeatScheduledClasses(ctx context.Context, user auth.User, base scheduledClassCreateParams, dates []string) ([]int64, error) {
	tx, err := dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	q := queries.New(tx)

	seriesID, err := q.InsertScheduledClassSeries(ctx, string(user.Role))
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	seriesNull := sql.NullInt64{Int64: seriesID, Valid: true}

	createdIDs := make([]int64, 0, len(dates))
	for _, date := range dates {
		params := base
		params.ScheduledDate = date
		params.SeriesID = seriesNull
		scheduleID, err := insertScheduledClassRow(ctx, q, user, params)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		createdIDs = append(createdIDs, scheduleID)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return createdIDs, nil
}
