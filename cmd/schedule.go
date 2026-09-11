package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/calendar"
	"zion-english/internal/classrules"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/meetings"
	"zion-english/internal/models"
	"zion-english/internal/notifications"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func handleSchedulePath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "schedule", "/edit/preview"); ok {
		handleScheduledClassEditPreview(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/edit/modal"); ok {
		handleScheduledClassEditModal(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/view"); ok {
		handleScheduledClassView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/conduct"); ok {
		handleConductScheduledClass(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/edit"); ok {
		handleScheduledClassEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/cancel"); ok {
		handleCancelScheduledClass(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/reschedule"); ok {
		handleRescheduleScheduledClass(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/delete"); ok {
		handleDeleteScheduledClass(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleSchedule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data := frontend.ScheduleData{}
		role := auth.GetRole(r.Context())
		if auth.IsTeacherScoped(role) {
			user := auth.GetUser(r.Context())
			data.LockTeacher = true
			data.TeacherID = strconv.FormatInt(user.ID, 10)
			data.TeacherName = user.Name
		} else {
			data.ShowAllTeachers = true
		}
		w.Header().Set("Content-Type", "text/html")
		frontend.Schedule(data).Render(r.Context(), w)
	case http.MethodPost:
		handleScheduleCreate(w, r)
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleScheduleRecord(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HttpRedirect(w, r, "/schedule")
	case http.MethodPost:
		handleScheduleCreate(w, r)
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleScheduleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	req, err := parseScheduledClassRequest(r, user, role)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.ScheduledDate,
		StartTime:       req.StartTime,
		DurationMinutes: req.DurationMinutes,
	}); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	var startTime sql.NullString
	if req.StartTime != "" {
		startTime = sql.NullString{String: req.StartTime, Valid: true}
	}

	scheduleID, err := dbRW.GetQueries().InsertScheduledClass(ctx, queries.InsertScheduledClassParams{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		ScheduledDate:   req.ScheduledDate,
		StartTime:       startTime,
		DurationMinutes: req.DurationMinutes,
		Rate:            req.Rate,
		Currency:        req.Currency,
		IsTrialClass:    trialClassToInt64(req.IsTrialClass),
		SeriesID:        sql.NullInt64{},
		Reason:          sql.NullString{},
		CreatedByRole:   string(user.Role),
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if err := saveScheduledClassLearningMaterials(ctx, user, scheduleID, parseLearningMaterialIDs(r)); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if meetingSvc != nil && meetings.SupportsAutoRoom(req.DurationMinutes) {
		student, studentErr := dbRO.GetQueries().GetStudentByID(ctx, req.StudentID)
		studentName := "student"
		if studentErr == nil {
			studentName = student.Name
		}
		if err := meetingSvc.SyncRoomForSchedule(ctx, meetings.ScheduledClassMeetingInput{
			ScheduleID:      scheduleID,
			TeacherID:       req.TeacherID,
			StudentName:     studentName,
			ScheduledDate:   req.ScheduledDate,
			StartTime:       req.StartTime,
			DurationMinutes: req.DurationMinutes,
		}); err != nil {
			logs.Log().Warn("zoom room sync failed after schedule create",
				zap.Error(err),
				zap.Int64("schedule_id", scheduleID),
				zap.Int64("teacher_id", req.TeacherID),
			)
		}
	}
	if req.StartTime != "" {
		syncCalendarForSchedule(ctx, scheduleID, req.TeacherID, req.StudentID, req.ScheduledDate, req.StartTime, req.DurationMinutes, req.Rate, req.Currency, "schedule create")
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "schedule", fmt.Sprintf("scheduled class for student id %d (teacher id %d, date %s)", req.StudentID, req.TeacherID, req.ScheduledDate))
	actor := auth.GetUser(ctx)
	notifyCrossParty(ctx, actor, req.TeacherID, teacherNameByID(ctx, req.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Class scheduled for %s", req.ScheduledDate))

	if err := respondFormMutation(w, "Class scheduled successfully!", "", "scheduleRefresh", "refreshScheduleCalendar"); err != nil {
		sendErrorLog(w, err.Error())
	}
}

func scheduledClassesCountParams(teacherID int64, startDate, endDate, statusFilter, nameFilter string) queries.CountScheduledClassesFilteredParams {
	return queries.CountScheduledClassesFilteredParams{
		Column1:         teacherID,
		TeacherID:       teacherID,
		ScheduledDate:   startDate,
		ScheduledDate_2: endDate,
		Column5:         statusFilter,
		Status:          statusFilter,
		Column7:         nameFilter,
		Column8:         sql.NullString{String: nameFilter, Valid: true},
	}
}

func scheduledClassesListParams(teacherID int64, startDate, endDate, statusFilter, nameFilter string, limit, offset int64) queries.GetScheduledClassesFilteredParams {
	return queries.GetScheduledClassesFilteredParams{
		Column1:         teacherID,
		TeacherID:       teacherID,
		ScheduledDate:   startDate,
		ScheduledDate_2: endDate,
		Column5:         statusFilter,
		Status:          statusFilter,
		Column7:         nameFilter,
		Column8:         sql.NullString{String: nameFilter, Valid: true},
		Limit:           limit,
		Offset:          offset,
	}
}

type scheduledClassesQuery struct {
	teacherID    int64
	startDate    string
	endDate      string
	statusFilter string
	nameFilter   string
}

func parseScheduledClassesQuery(r *http.Request) (scheduledClassesQuery, error) {
	teacherIDStr := r.URL.Query().Get("teacherId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	statusFilter := r.URL.Query().Get("status")
	nameFilter := r.URL.Query().Get("q")

	if startDate == "" || endDate == "" {
		return scheduledClassesQuery{}, errors.New("missing required parameters")
	}

	role := auth.GetRole(r.Context())
	var teacherID int64
	if teacherIDStr == "" || teacherIDStr == "0" {
		if !auth.HasAdminAccess(role) {
			return scheduledClassesQuery{}, errors.New("missing required parameters")
		}
	} else {
		parsedID, err := strconv.ParseInt(teacherIDStr, 10, 64)
		if err != nil {
			return scheduledClassesQuery{}, errors.New("invalid teacher ID")
		}
		teacherID = parsedID
		if auth.IsTeacherScoped(role) {
			user := auth.GetUser(r.Context())
			if teacherID != user.ID {
				return scheduledClassesQuery{}, errors.New("forbidden")
			}
		}
	}

	return scheduledClassesQuery{
		teacherID:    teacherID,
		startDate:    startDate,
		endDate:      endDate,
		statusFilter: statusFilter,
		nameFilter:   nameFilter,
	}, nil
}

func fetchScheduledClassViews(ctx context.Context, q scheduledClassesQuery, limit, offset int64) ([]models.ScheduledClassView, error) {
	records, err := dbRO.GetQueries().GetScheduledClassesFiltered(ctx, scheduledClassesListParams(q.teacherID, q.startDate, q.endDate, q.statusFilter, q.nameFilter, limit, offset))
	if err != nil {
		return nil, err
	}

	classIDs := make([]int64, 0, len(records))
	for _, sc := range records {
		classIDs = append(classIDs, sc.ID)
	}
	roomMap := map[int64]meetings.ClassMeetingView{}
	if meetingSvc != nil && len(classIDs) > 0 {
		rooms, roomErr := meetingSvc.GetRoomsByClassIDs(ctx, classIDs)
		if roomErr == nil {
			roomMap = rooms
		}
	}
	calendarMap := map[int64]calendar.CalendarEventView{}
	if calendarSvc != nil && len(classIDs) > 0 {
		events, calendarErr := calendarSvc.GetEventsByClassIDs(ctx, classIDs)
		if calendarErr == nil {
			calendarMap = events
		}
	}

	response := make([]models.ScheduledClassView, 0, len(records))
	teacherIDs := make([]int64, 0, len(records))
	for _, sc := range records {
		view := scheduledClassViewFromRow(sc)
		if room, ok := roomMap[sc.ID]; ok {
			view.RoomURL = room.RoomURL
			view.RoomPasscode = room.Passcode
			view.MeetingService = room.Service
		}
		if event, ok := calendarMap[sc.ID]; ok {
			view.CalendarEventURL = event.EventURL
			view.CalendarService = event.Service
		}
		response = append(response, view)
		teacherIDs = append(teacherIDs, sc.TeacherID)
	}

	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, err
	}
	enrichScheduledClassViewsWithRoleBadges(response, rolesMap)
	return response, nil
}

func handleScheduleListPartial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q, err := parseScheduledClassesQuery(r)
	if err != nil {
		if err.Error() == "forbidden" {
			HttpError(w, err.Error(), http.StatusForbidden)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	views, err := fetchScheduledClassViews(r.Context(), q, 500, 0)
	if err != nil {
		HttpError(w, "Failed to fetch scheduled classes", http.StatusInternalServerError)
		return
	}

	items := frontend.ScheduledClassItemsFromViews(views)
	items, err = enrichScheduledClassItemsSeries(r.Context(), items)
	if err != nil {
		HttpError(w, "Failed to fetch scheduled classes", http.StatusInternalServerError)
		return
	}
	emptyMsg := "No classes scheduled for this day."
	if q.startDate == q.endDate && q.startDate == utils.TodayPHT() {
		emptyMsg = "No classes scheduled for today."
	}
	role := auth.GetRole(r.Context())
	if auth.HasAdminAccess(role) && q.teacherID == 0 {
		teacherParam := strings.TrimSpace(r.URL.Query().Get("teacherId"))
		if teacherParam == "" || teacherParam == "0" {
			if q.startDate == utils.TodayPHT() && q.startDate == q.endDate {
				emptyMsg = "Select a teacher to view today's schedule."
			}
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ScheduledClassList(items, emptyMsg, utils.URL("/static/zoom-logo.svg"), utils.URL("/static/google-calendar-logo.svg")).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleScheduleDayTimelinePartial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q, err := parseScheduledClassesQuery(r)
	if err != nil {
		if err.Error() == "forbidden" {
			HttpError(w, err.Error(), http.StatusForbidden)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	views, err := fetchScheduledClassViews(r.Context(), q, 500, 0)
	if err != nil {
		HttpError(w, "Failed to fetch scheduled classes", http.StatusInternalServerError)
		return
	}

	items := frontend.ScheduledClassItemsFromViews(views)
	items, err = enrichScheduledClassItemsSeries(r.Context(), items)
	if err != nil {
		HttpError(w, "Failed to fetch scheduled classes", http.StatusInternalServerError)
		return
	}
	emptyMsg := "No classes scheduled for this day."
	if q.startDate == q.endDate && q.startDate == utils.TodayPHT() {
		emptyMsg = "No classes scheduled for today."
	}
	role := auth.GetRole(r.Context())
	if auth.HasAdminAccess(role) && q.teacherID == 0 {
		teacherParam := strings.TrimSpace(r.URL.Query().Get("teacherId"))
		if teacherParam == "" || teacherParam == "0" {
			emptyMsg = "Select a teacher to view the day schedule."
		}
	}

	timeline := frontend.BuildScheduledClassDayTimeline(items, emptyMsg)
	timeline.ZoomLogoURL = utils.URL("/static/zoom-logo.svg")
	timeline.CalendarLogoURL = utils.URL("/static/google-calendar-logo.svg")
	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ScheduledClassDayTimeline(timeline).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGetScheduledClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q, err := parseScheduledClassesQuery(r)
	if err != nil {
		if err.Error() == "forbidden" {
			HttpError(w, err.Error(), http.StatusForbidden)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	page := utils.ParsePageQuery(r)
	ctx := r.Context()

	countParams := scheduledClassesCountParams(q.teacherID, q.startDate, q.endDate, q.statusFilter, q.nameFilter)
	total, err := dbRO.GetQueries().CountScheduledClassesFiltered(ctx, countParams)
	if err != nil {
		HttpError(w, "Failed to count scheduled classes", http.StatusInternalServerError)
		return
	}
	page.Total = total

	response, err := fetchScheduledClassViews(ctx, q, int64(page.Size), int64(page.Offset()))
	if err != nil {
		HttpError(w, "Failed to fetch scheduled classes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": response,
		"page": map[string]interface{}{
			"number":     page.Number,
			"size":       page.Size,
			"total":      page.Total,
			"totalPages": page.TotalPages(),
		},
	})
}

func handleCancelScheduledClass(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		HttpError(w, "Scheduled class not found", http.StatusNotFound)
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if existing.Status != "scheduled" {
		sendErrorLog(w, "only scheduled classes can be cancelled")
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err := validateClassReason(reason); err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	notes := strings.TrimSpace(r.FormValue("notes"))
	scope := constants.ParseSeriesScope(r.FormValue("series_scope"))
	targets, err := scheduledClassesForScope(ctx, existing, scope)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	for _, target := range targets {
		if err := cancelScheduledClassByID(ctx, user, target.ID, reason, notes); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	notifyCrossParty(ctx, user, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Scheduled class%s cancelled", formatSeriesScopeSummary(int64(len(targets)), "cancelled")))

	message := "Class cancelled."
	if len(targets) > 1 {
		message = fmt.Sprintf("%d classes cancelled.", len(targets))
	}
	respondScheduledClassAction(w, r.FormValue("from"), message)
}

func handleRescheduleScheduledClass(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		HttpError(w, "Scheduled class not found", http.StatusNotFound)
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if existing.Status != "scheduled" {
		sendErrorLog(w, "only scheduled classes can be rescheduled")
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	newDate := r.FormValue("scheduled_date")
	if newDate == "" {
		sendErrorLog(w, "new date is required")
		return
	}
	newTime := r.FormValue("start_time")
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err := validateClassReason(reason); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	effectiveStart := normalizeScheduleStartTime(newTime)
	if effectiveStart == "" && existing.StartTime.Valid {
		effectiveStart = existing.StartTime.String
	}

	if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
		ScheduleID:      scheduleID,
		StudentID:       existing.StudentID,
		TeacherID:       existing.TeacherID,
		Date:            newDate,
		StartTime:       effectiveStart,
		DurationMinutes: existing.DurationMinutes,
	}); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	var startTime sql.NullString
	if newTime != "" {
		startTime = sql.NullString{String: newTime, Valid: true}
	} else if existing.StartTime.Valid {
		startTime = existing.StartTime
	}

	err = dbRW.GetQueries().RescheduleScheduledClass(ctx, queries.RescheduleScheduledClassParams{
		ScheduledDate: newDate,
		StartTime:     startTime,
		Reason:        sql.NullString{String: reason, Valid: true},
		ID:            scheduleID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if meetingSvc != nil {
		student, studentErr := dbRO.GetQueries().GetStudentByID(ctx, existing.StudentID)
		studentName := "student"
		if studentErr == nil {
			studentName = student.Name
		}
		if err := meetingSvc.SyncRoomForSchedule(ctx, meetings.ScheduledClassMeetingInput{
			ScheduleID:      scheduleID,
			TeacherID:       existing.TeacherID,
			StudentName:     studentName,
			ScheduledDate:   newDate,
			StartTime:       effectiveStart,
			DurationMinutes: existing.DurationMinutes,
		}); err != nil {
			logs.Log().Warn("zoom room sync failed after reschedule",
				zap.Error(err),
				zap.Int64("schedule_id", scheduleID),
				zap.Int64("teacher_id", existing.TeacherID),
			)
		}
	}
	if effectiveStart != "" {
		syncCalendarForSchedule(ctx, scheduleID, existing.TeacherID, existing.StudentID, newDate, effectiveStart, existing.DurationMinutes, existing.Rate, existing.Currency, "reschedule")
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "schedule", fmt.Sprintf("rescheduled class id %d from %s to %s", scheduleID, existing.ScheduledDate, newDate))
	actor := auth.GetUser(ctx)
	notifyCrossParty(ctx, actor, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Scheduled class moved from %s to %s", existing.ScheduledDate, newDate))

	if _, err := fmt.Fprint(w, "Class rescheduled.\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
}

func handleDeleteScheduledClass(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	reason := strings.TrimSpace(r.FormValue("reason"))
	if err := validateDeletionReason(reason); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		HttpError(w, "Scheduled class not found", http.StatusNotFound)
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if existing.Status != "scheduled" {
		sendErrorLog(w, "only scheduled classes can be deleted")
		return
	}

	scope := constants.ParseSeriesScope(r.FormValue("series_scope"))
	targets, err := scheduledClassesForScope(ctx, existing, scope)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	for _, target := range targets {
		if err := deleteScheduledClassByID(ctx, user, target.ID, reason); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	notifyCrossParty(ctx, user, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Scheduled class%s deleted", formatSeriesScopeSummary(int64(len(targets)), "deleted")))

	message := "Scheduled class deleted successfully."
	if len(targets) > 1 {
		message = fmt.Sprintf("%d scheduled classes deleted successfully.", len(targets))
	}
	respondScheduledClassAction(w, r.FormValue("from"), message)
}

func editScheduleData(ctx context.Context, scheduleID int64, lockTeacher, isSuperuser bool) (frontend.EditScheduleData, error) {
	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		return frontend.EditScheduleData{}, err
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, auth.GetUser(ctx)); err != nil {
		return frontend.EditScheduleData{}, err
	}

	startTime := ""
	if existing.StartTime.Valid {
		startTime = existing.StartTime.String
	}

	materials, err := loadScheduledClassLearningMaterialLinks(ctx, scheduleID)
	if err != nil {
		return frontend.EditScheduleData{}, err
	}

	seriesID, _ := seriesIDFromRow(existing.SeriesID)
	total, future, err := seriesCounts(ctx, existing.SeriesID, existing.ScheduledDate)
	if err != nil {
		return frontend.EditScheduleData{}, err
	}
	dates, err := seriesDates(ctx, existing.SeriesID)
	if err != nil {
		return frontend.EditScheduleData{}, err
	}

	return frontend.EditScheduleData{
		ScheduleID:  strconv.FormatInt(scheduleID, 10),
		LockTeacher: lockTeacher,
		IsSuperuser: isSuperuser,
		TeacherID:   strconv.FormatInt(existing.TeacherID, 10),
		TeacherName: existing.TeacherName,
		StudentID:   strconv.FormatInt(existing.StudentID, 10),
		Date:        existing.ScheduledDate,
		StartTime:   startTime,
		EndTime:     utils.EndTimeFromStartAndDuration(startTime, existing.DurationMinutes),
		Rate:              existing.Rate,
		Currency:          existing.Currency,
		IsTrialClass:      existing.IsTrialClass != 0,
		LearningMaterials: materials,
		SeriesID:          seriesID,
		SeriesTotalCount:  total,
		SeriesFutureCount: future,
		SeriesDates:       dates,
	}, nil
}

func handleScheduledClassEdit(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	ctx := r.Context()
	user := auth.GetUser(ctx)

	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		HttpError(w, "Scheduled class not found", http.StatusNotFound)
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if existing.Status != "scheduled" {
		HttpError(w, "only scheduled classes can be edited", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		HttpRedirect(w, r, "/schedule")
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	studentID, err := formInt64(r, "schedule_student", "student")
	if err != nil {
		sendErrorLog(w, "student is required")
		return
	}

	startTime := normalizeScheduleStartTime(r.FormValue("start_time"))
	endTime := r.FormValue("end_time")
	duration, err := utils.DurationMinutesFromRange(startTime, endTime)
	if err != nil {
		sendErrorLog(w, friendlyTimeRangeError(err).Error())
		return
	}

	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	detailsReq := models.ScheduledClassRequest{
		StudentID:    studentID,
		Rate:         rate,
		Currency:     r.FormValue("currency"),
		IsTrialClass: formIsTrialClass(r),
	}
	models.ApplyScheduledTrialClassRate(&detailsReq)
	if err := validateScheduledClassRateCurrency(detailsReq.Rate, detailsReq.Currency); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	scope := constants.ParseSeriesScope(r.FormValue("edit_scope"))
	lmIDs := parseLearningMaterialIDs(r)
	editInput := scheduledClassEditInput{
		StudentID:       studentID,
		StartTime:       startTime,
		DurationMinutes: duration,
		Rate:            detailsReq.Rate,
		Currency:        detailsReq.Currency,
		IsTrialClass:    detailsReq.IsTrialClass,
	}

	existingStart := ""
	if existing.StartTime.Valid {
		existingStart = existing.StartTime.String
	}

	if _, ok := seriesIDFromRow(existing.SeriesID); ok {
		allSeriesDates, err := seriesDates(ctx, existing.SeriesID)
		if err != nil {
			sendErrorLog(w, err.Error())
			return
		}
		confirmDates, err := parseEditSeriesConfirmDates(r, allSeriesDates)
		if err != nil {
			sendErrorLog(w, err.Error())
			return
		}
		if confirmDates != nil {
			affected, err := applySeriesScheduleEdit(ctx, user, existing, scope, confirmDates, allSeriesDates, editInput, lmIDs, existingStart)
			if err != nil {
				sendErrorLog(w, err.Error())
				return
			}
			auditAction := "updated scheduled class"
			if affected > 1 {
				auditAction = fmt.Sprintf("updated %d scheduled classes in series", affected)
			}
			insertAuditLogAs(ctx, user, "schedule", fmt.Sprintf("%s starting id %d (student id %d)", auditAction, scheduleID, studentID))
			notifyCrossParty(ctx, user, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
				fmt.Sprintf("Scheduled class%s updated", formatSeriesScopeSummary(int64(affected), "updated")))
			respondScheduledClassAction(w, r.FormValue("from"), "Scheduled class updated successfully!")
			return
		}
	}

	newDate := r.FormValue("scheduled_date")
	if newDate == "" {
		newDate = r.FormValue("date")
	}
	if newDate == "" {
		sendErrorLog(w, "date is required")
		return
	}
	if _, err := utils.ParseDatePHT(newDate); err != nil {
		sendErrorLog(w, "invalid date format")
		return
	}

	targets, err := scheduledClassesForScope(ctx, existing, scope)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if scope == constants.SeriesScopeSingle {
		if err := validateScheduleDateTimeChange(existing.ScheduledDate, existingStart, newDate, startTime); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	} else if newDate != existing.ScheduledDate {
		sendErrorLog(w, "date cannot be changed when updating this and future classes in a series")
		return
	}

	for _, target := range targets {
		targetDate := target.ScheduledDate
		if scope == constants.SeriesScopeSingle {
			targetDate = newDate
		}
		editInput.ScheduledDate = targetDate
		if err := applyScheduledClassEdit(ctx, user, target, editInput, lmIDs); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	auditAction := "updated scheduled class"
	if len(targets) > 1 {
		auditAction = fmt.Sprintf("updated %d scheduled classes in series", len(targets))
	}
	insertAuditLogAs(ctx, user, "schedule", fmt.Sprintf("%s starting id %d (student id %d)", auditAction, scheduleID, studentID))
	notifyCrossParty(ctx, user, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Scheduled class%s updated", formatSeriesScopeSummary(int64(len(targets)), "updated")))

	respondScheduledClassAction(w, r.FormValue("from"), "Scheduled class updated successfully!")
}

func handleScheduledClassEditPreview(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		HttpError(w, "Scheduled class not found", http.StatusNotFound)
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if existing.Status != "scheduled" {
		HttpError(w, "only scheduled classes can be edited", http.StatusBadRequest)
		return
	}

	if _, ok := seriesIDFromRow(existing.SeriesID); !ok {
		HttpError(w, "this class is not part of a repeating series", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	studentID, err := formInt64(r, "schedule_student", "student")
	if err != nil {
		sendErrorLog(w, "student is required")
		return
	}

	startTime := normalizeScheduleStartTime(r.FormValue("start_time"))
	endTime := r.FormValue("end_time")
	duration, err := utils.DurationMinutesFromRange(startTime, endTime)
	if err != nil {
		sendErrorLog(w, friendlyTimeRangeError(err).Error())
		return
	}

	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	detailsReq := models.ScheduledClassRequest{
		StudentID:    studentID,
		Rate:         rate,
		Currency:     r.FormValue("currency"),
		IsTrialClass: formIsTrialClass(r),
	}
	models.ApplyScheduledTrialClassRate(&detailsReq)
	if err := validateScheduledClassRateCurrency(detailsReq.Rate, detailsReq.Currency); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	allSeriesDates, err := seriesDates(ctx, existing.SeriesID)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	selectedDates, err := parseEditSeriesDates(r, allSeriesDates)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if selectedDates == nil {
		sendErrorLog(w, "select at least one date for this series")
		return
	}

	scope := constants.ParseSeriesScope(r.FormValue("edit_scope"))
	editInput := scheduledClassEditInput{
		StudentID:       studentID,
		StartTime:       startTime,
		DurationMinutes: duration,
		Rate:            detailsReq.Rate,
		Currency:        detailsReq.Currency,
		IsTrialClass:    detailsReq.IsTrialClass,
	}
	rows, err := previewSeriesEditRows(ctx, user, existing, scope, selectedDates, editInput)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ScheduleSeriesEditPreview(frontend.ScheduleSeriesEditPreviewData{
		ScheduleID: strconv.FormatInt(scheduleID, 10),
		Scope:      string(scope),
		From:       r.FormValue("from"),
		LockTeacher: auth.IsTeacherScoped(role),
		Base: frontend.ScheduleRepeatFormSnapshot{
			TeacherID:    strconv.FormatInt(existing.TeacherID, 10),
			StudentID:    strconv.FormatInt(studentID, 10),
			StartTime:    startTime,
			EndTime:      endTime,
			Rate:         detailsReq.Rate,
			Currency:     detailsReq.Currency,
			IsTrialClass: detailsReq.IsTrialClass,
		},
		Rows:                rows,
		LearningMaterialIDs: parseLearningMaterialIDs(r),
	}).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleScheduledClassEditModal(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		HttpError(w, "Scheduled class not found", http.StatusNotFound)
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if existing.Status != "scheduled" {
		HttpError(w, "only scheduled classes can be edited", http.StatusBadRequest)
		return
	}

	data, err := editScheduleData(ctx, scheduleID, auth.IsTeacherScoped(role), auth.HasAdminAccess(role))
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}
	data.TodayPHT = utils.TodayPHT()
	data.From = r.URL.Query().Get("from")

	w.Header().Set("Content-Type", "text/html")
	frontend.EditScheduledClassModalForm(data).Render(ctx, w)
}

func handleScheduledClassView(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	data, err := scheduledClassViewData(ctx, scheduleID)
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.ClassViewModal(data).Render(ctx, w)
}

func scheduledClassViewData(ctx context.Context, scheduleID int64) (frontend.EditClassData, error) {
	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		return frontend.EditClassData{}, err
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, auth.GetUser(ctx)); err != nil {
		return frontend.EditClassData{}, err
	}

	teacher, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, existing.TeacherID)
	if err != nil {
		return frontend.EditClassData{}, err
	}

	teacherRoles, err := loadTeacherRoles(ctx, existing.TeacherID)
	if err != nil {
		return frontend.EditClassData{}, err
	}

	materials, err := loadScheduledClassLearningMaterialLinks(ctx, scheduleID)
	if err != nil {
		return frontend.EditClassData{}, err
	}

	startTime := ""
	if existing.StartTime.Valid {
		startTime = existing.StartTime.String
	}
	reason := ""
	if existing.Reason.Valid {
		reason = existing.Reason.String
	}

	role := auth.GetRole(ctx)
	seriesID, _ := seriesIDFromRow(existing.SeriesID)
	_, future, err := seriesCounts(ctx, existing.SeriesID, existing.ScheduledDate)
	if err != nil {
		return frontend.EditClassData{}, err
	}
	return frontend.EditClassData{
		RecordID:        strconv.FormatInt(scheduleID, 10),
		Readonly:        true,
		IsSuperuser:     auth.HasAdminAccess(role),
		StudentID:       strconv.FormatInt(existing.StudentID, 10),
		TeacherID:       strconv.FormatInt(existing.TeacherID, 10),
		StudentName:     existing.StudentName,
		TeacherName:     existing.TeacherName,
		TeacherAvatar: avatarWithTeacherRoles(
			buildTeacherListAvatarProps(
				teacher.ID,
				teacher.FirstName,
				teacher.MiddleName,
				teacher.LastName,
				teacher.AssignedColor,
				teacher.ProfilePicture,
			),
			teacherRoles,
		),
		Date:              existing.ScheduledDate,
		StartTime:         startTime,
		EndTime:           utils.EndTimeFromStartAndDuration(startTime, existing.DurationMinutes),
		DurationMinutes:   existing.DurationMinutes,
		Rate:              existing.Rate,
		Currency:          existing.Currency,
		IsTrialClass:      existing.IsTrialClass != 0,
		Status:            constants.ClassListFilterScheduled,
		Reason:            reason,
		LearningMaterials: materials,
		ActionFrom:        frontend.ClassActionContextSchedule,
		SeriesID:          seriesID,
		SeriesFutureCount: future,
	}, nil
}

func handleConductScheduledClass(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		HttpError(w, "Scheduled class not found", http.StatusNotFound)
		return
	}

	rules := classrules.ScheduledClassRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if existing.Status != "scheduled" {
		sendErrorLog(w, "only scheduled classes can be conducted")
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}
	notes := strings.TrimSpace(r.FormValue("notes"))

	if err := insertClassRecordFromSchedule(ctx, user, existing, scheduleID, "conducted", "", notes); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLogAs(ctx, user, "schedule", fmt.Sprintf("marked scheduled class id %d as conducted", scheduleID))
	insertAuditLogAs(ctx, user, "classes", fmt.Sprintf("recorded class for student id %d (teacher id %d, date %s, status conducted)", existing.StudentID, existing.TeacherID, existing.ScheduledDate))
	notifyCrossParty(ctx, user, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindClassRecorded,
		fmt.Sprintf("Class recorded for student on %s (status conducted)", existing.ScheduledDate))

	respondScheduledClassAction(w, r.FormValue("from"), "Class conducted successfully!")
}

func classRecordRequestFromSchedule(existing queries.GetScheduledClassByIDRow, status, reason, notes string) models.ClassRecordRequest {
	startTime := ""
	if existing.StartTime.Valid {
		startTime = existing.StartTime.String
	}
	endTime := utils.EndTimeFromStartAndDuration(startTime, existing.DurationMinutes)
	req := models.ClassRecordRequest{
		StudentID:       existing.StudentID,
		TeacherID:       existing.TeacherID,
		Date:            existing.ScheduledDate,
		StartTime:       startTime,
		EndTime:         endTime,
		DurationMinutes: existing.DurationMinutes,
		Rate:            existing.Rate,
		Currency:        existing.Currency,
		IsTrialClass:    existing.IsTrialClass != 0,
		Status:          status,
		Reason:          reason,
		Notes:           notes,
	}
	return req
}

func insertClassRecordFromSchedule(ctx context.Context, user auth.User, existing queries.GetScheduledClassByIDRow, scheduleID int64, status, reason, notes string) error {
	req := classRecordRequestFromSchedule(existing, status, reason, notes)
	if err := validateClassRecordRequest(&req); err != nil {
		return err
	}
	if err := applyClassRecordRules(ctx, user, req, 0); err != nil {
		return err
	}

	recordID, err := dbRW.GetQueries().InsertClassRecord(ctx, queries.InsertClassRecordParams{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.Date,
		StartTime:       sql.NullString{String: req.StartTime, Valid: req.StartTime != ""},
		EndTime:         sql.NullString{String: req.EndTime, Valid: req.EndTime != ""},
		DurationMinutes: req.DurationMinutes,
		Rate:            req.Rate,
		Currency:        req.Currency,
		IsTrialClass:    trialClassToInt64(req.IsTrialClass),
		Status:          req.Status,
		Reason:          sql.NullString{String: req.Reason, Valid: req.Reason != ""},
		Notes:           sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		RecordedByRole:  string(user.Role),
	})
	if err != nil {
		return err
	}

	if err := copyScheduledClassLearningMaterials(ctx, dbRW.GetQueries(), scheduleID, recordID); err != nil {
		return err
	}

	return markScheduledClassConducted(ctx, scheduleID, req)
}

func validateScheduleDateTimeChange(existingDate, existingStart, newDate, newStart string) error {
	if newDate == existingDate && normalizeScheduleStartTime(newStart) == normalizeScheduleStartTime(existingStart) {
		return nil
	}
	if utils.IsDateTimeInPastPHT(newDate, newStart) {
		return errors.New("date and time cannot be in the past")
	}
	return nil
}

func respondScheduledClassAction(w http.ResponseWriter, from, message string) {
	setSuccessFlash(w, message)
	switch from {
	case "schedule":
		w.Header().Set("HX-Trigger", `{"scheduleDayOpen":null,"refreshScheduleCalendar":null}`)
	case string(frontend.ClassActionContextScheduleSeries):
		w.Header().Set("HX-Trigger", "scheduleSeriesRefresh")
	default:
		setListRefreshTriggers(w, "classesRefresh")
	}
	if _, err := fmt.Fprint(w, message+"\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
}

func markScheduledClassConducted(ctx context.Context, scheduleID int64, req models.ClassRecordRequest) error {
	existing, err := dbRO.GetQueries().GetScheduledClassByID(ctx, scheduleID)
	if err != nil {
		return errors.New("scheduled class not found")
	}

	if existing.Status != "scheduled" {
		return errors.New("scheduled class is not in scheduled status")
	}
	if existing.StudentID != req.StudentID ||
		existing.TeacherID != req.TeacherID ||
		existing.ScheduledDate != req.Date ||
		existing.DurationMinutes != req.DurationMinutes {
		return errors.New("class record does not match scheduled class")
	}

	return dbRW.GetQueries().UpdateScheduledClassStatus(ctx, queries.UpdateScheduledClassStatusParams{
		Status: req.Status,
		Reason: sql.NullString{String: req.Reason, Valid: req.Status != "conducted" && req.Reason != ""},
		ID:     scheduleID,
	})
}

func parseScheduledClassRequest(r *http.Request, user auth.User, role auth.Role) (models.ScheduledClassRequest, error) {
	studentID, err := formInt64(r, "schedule_student", "student")
	if err != nil {
		return models.ScheduledClassRequest{}, errors.New("student is required")
	}
	startTime := r.FormValue("start_time")
	endTime := r.FormValue("end_time")
	duration, err := utils.DurationMinutesFromRange(startTime, endTime)
	if err != nil {
		return models.ScheduledClassRequest{}, friendlyTimeRangeError(err)
	}
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		return models.ScheduledClassRequest{}, err
	}

	teacherID := user.ID
	if auth.HasAdminAccess(role) {
		teacherID, err = formInt64(r, "schedule_teacher", "teacher")
		if err != nil {
			return models.ScheduledClassRequest{}, errors.New("teacher is required")
		}
	}

	req := models.ScheduledClassRequest{
		StudentID:       studentID,
		TeacherID:       teacherID,
		ScheduledDate:   r.FormValue("scheduled_date"),
		StartTime:       normalizeScheduleStartTime(r.FormValue("start_time")),
		DurationMinutes: duration,
		Rate:            rate,
		Currency:        r.FormValue("currency"),
		IsTrialClass:    formIsTrialClass(r),
	}
	models.ApplyScheduledTrialClassRate(&req)
	return req, validateScheduledClassRequest(&req)
}

func normalizeScheduleStartTime(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := utils.ParseTimeHM(value)
	if err != nil {
		return value
	}
	return parsed.Format("15:04")
}

func formInt64(r *http.Request, names ...string) (int64, error) {
	for _, name := range names {
		if value := strings.TrimSpace(r.FormValue(name)); value != "" {
			return requireInt64(value)
		}
	}
	return 0, errors.New("missing integer value")
}

func validateScheduledClassRateCurrency(rate float64, currency string) error {
	if rate <= 0 {
		return errors.New("rate must be greater than zero")
	}
	if currency == "" {
		return errors.New("currency is required")
	}
	if !constants.ValidCurrency(currency) {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}
	return nil
}

func validateScheduledClassRequest(req *models.ScheduledClassRequest) error {
	if req.StudentID == 0 {
		return errors.New("student is required")
	}
	if req.TeacherID == 0 {
		return errors.New("teacher is required")
	}
	if req.ScheduledDate == "" {
		return errors.New("date is required")
	}
	if _, err := utils.ParseDatePHT(req.ScheduledDate); err != nil {
		return errors.New("invalid date format")
	}
	if req.StartTime == "" {
		return errors.New("start time is required")
	}
	if _, err := utils.ParseTimeHM(req.StartTime); err != nil {
		return utils.ErrInvalidStartTime
	}
	if req.DurationMinutes <= 0 {
		return errors.New("duration must be greater than zero")
	}
	return validateScheduledClassRateCurrency(req.Rate, req.Currency)
}

func scheduledClassTeacherAvatar(sc queries.GetScheduledClassesFilteredRow) models.AvatarView {
	hasPicture := sc.TeacherProfilePicture.Valid && sc.TeacherProfilePicture.String != ""
	assignedColor := sc.TeacherAssignedColor
	if assignedColor == "" {
		assignedColor = constants.DefaultTeacherAssignedColor
	}
	return models.AvatarView{
		Initials:      utils.PersonInitials(sc.TeacherFirstName, sc.TeacherMiddleName, sc.TeacherLastName, sc.TeacherName),
		AssignedColor: assignedColor,
		HasPicture:    hasPicture,
		PictureURL:    teacherPictureURL(sc.TeacherID, hasPicture),
		Alt:           sc.TeacherName + " avatar",
	}
}

func scheduledClassViewFromRow(sc queries.GetScheduledClassesFilteredRow) models.ScheduledClassView {
	startTime := ""
	if sc.StartTime.Valid {
		startTime = sc.StartTime.String
	}
	reason := ""
	if sc.Reason.Valid {
		reason = sc.Reason.String
	}
	return models.ScheduledClassView{
		ID:              sc.ID,
		StudentID:       sc.StudentID,
		TeacherID:       sc.TeacherID,
		StudentName:     sc.StudentName,
		TeacherName:     sc.TeacherName,
		TeacherAvatar:   scheduledClassTeacherAvatar(sc),
		ScheduledDate:   sc.ScheduledDate,
		StartTime:       startTime,
		EndTime:         utils.EndTimeFromStartAndDuration(startTime, sc.DurationMinutes),
		DurationMinutes: sc.DurationMinutes,
		Rate:            sc.Rate,
		Currency:        sc.Currency,
		Status:          sc.Status,
		Reason:          reason,
		CreatedAt:       sc.CreatedAt,
		SeriesID:        scheduledClassSeriesID(sc.SeriesID),
	}
}

func scheduledClassSeriesID(seriesID sql.NullInt64) int64 {
	if !seriesID.Valid {
		return 0
	}
	return seriesID.Int64
}

func syncCalendarForSchedule(ctx context.Context, scheduleID, teacherID, studentID int64, scheduledDate, startTime string, durationMinutes int64, rate float64, currency, action string) {
	if calendarSvc == nil || strings.TrimSpace(startTime) == "" {
		return
	}
	student, studentErr := dbRO.GetQueries().GetStudentByID(ctx, studentID)
	studentName := "student"
	if studentErr == nil {
		studentName = student.Name
	}
	if err := calendarSvc.SyncEventForSchedule(ctx, calendar.ScheduledClassCalendarInput{
		ScheduleID:      scheduleID,
		TeacherID:       teacherID,
		StudentName:     studentName,
		ScheduledDate:   scheduledDate,
		StartTime:       startTime,
		DurationMinutes: durationMinutes,
		Rate:            rate,
		Currency:        currency,
	}); err != nil {
		logs.Log().Warn("google calendar sync failed after "+action,
			zap.Error(err),
			zap.Int64("schedule_id", scheduleID),
			zap.Int64("teacher_id", teacherID),
		)
	}
}
