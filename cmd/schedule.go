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
		if role == auth.RoleTeacher {
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
		Reason:          sql.NullString{},
		CreatedByRole:   string(user.Role),
	})
	if err != nil {
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

	insertAuditLogAs(ctx, auth.GetUser(ctx), "schedule", fmt.Sprintf("scheduled class for student id %d (teacher id %d, date %s)", req.StudentID, req.TeacherID, req.ScheduledDate))
	actor := auth.GetUser(ctx)
	notifyCrossParty(ctx, actor, req.TeacherID, teacherNameByID(ctx, req.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Class scheduled for %s", req.ScheduledDate))

	if _, err := fmt.Fprint(w, "Class scheduled successfully!\n"); err != nil {
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
		if role != auth.RoleSuperuser {
			return scheduledClassesQuery{}, errors.New("missing required parameters")
		}
	} else {
		parsedID, err := strconv.ParseInt(teacherIDStr, 10, 64)
		if err != nil {
			return scheduledClassesQuery{}, errors.New("invalid teacher ID")
		}
		teacherID = parsedID
		if role == auth.RoleTeacher {
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

	response := make([]models.ScheduledClassView, 0, len(records))
	for _, sc := range records {
		view := scheduledClassViewFromRow(sc)
		if room, ok := roomMap[sc.ID]; ok {
			view.RoomURL = room.RoomURL
			view.RoomPasscode = room.Passcode
			view.MeetingService = room.Service
		}
		response = append(response, view)
	}
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
	emptyMsg := "No classes scheduled for this day."
	if q.startDate == q.endDate && q.startDate == utils.TodayPHT() {
		emptyMsg = "No classes scheduled for today."
	}
	role := auth.GetRole(r.Context())
	if role == auth.RoleSuperuser && q.teacherID == 0 {
		teacherParam := strings.TrimSpace(r.URL.Query().Get("teacherId"))
		if teacherParam == "" || teacherParam == "0" {
			if q.startDate == utils.TodayPHT() && q.startDate == q.endDate {
				emptyMsg = "Select a teacher to view today's schedule."
			}
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ScheduledClassList(items, emptyMsg, utils.URL("/static/zoom-logo.svg")).Render(r.Context(), w); err != nil {
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

	if meetingSvc != nil {
		_ = meetingSvc.DeleteRoomForSchedule(ctx, scheduleID, existing.TeacherID)
	}

	err = dbRW.GetQueries().UpdateScheduledClassStatus(ctx, queries.UpdateScheduledClassStatusParams{
		Status:   "cancelled",
		Reason:   sql.NullString{String: reason, Valid: true},
		ID:       scheduleID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "schedule", fmt.Sprintf("cancelled scheduled class id %d (student id %d, date %s)", scheduleID, existing.StudentID, existing.ScheduledDate))
	actor := auth.GetUser(ctx)
	notifyCrossParty(ctx, actor, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Scheduled class on %s was cancelled", existing.ScheduledDate))

	if _, err := fmt.Fprint(w, "Class cancelled.\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
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

	if meetingSvc != nil {
		_ = meetingSvc.DeleteRoomForSchedule(ctx, scheduleID, existing.TeacherID)
	}

	err = dbRW.GetQueries().SoftDeleteScheduledClass(ctx, queries.SoftDeleteScheduledClassParams{
		Reason: sql.NullString{String: reason, Valid: true},
		ID:     scheduleID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLogAs(ctx, user, "schedule", fmt.Sprintf("deleted scheduled class id %d (student id %d, date %s, reason: %s)", scheduleID, existing.StudentID, existing.ScheduledDate, reason))
	notifyCrossParty(ctx, user, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Scheduled class on %s was deleted", existing.ScheduledDate))

	from := r.FormValue("from")
	setSuccessFlash(w, "Scheduled class deleted successfully.")
	if from == "schedule" {
		w.Header().Set("HX-Trigger", "scheduleDayOpen")
		if _, err := fmt.Fprint(w, "Scheduled class deleted.\n"); err != nil {
			sendErrorLog(w, err.Error())
		}
		return
	}
	HttpRedirect(w, r, "/classes")
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
		Rate:        existing.Rate,
		Currency:    existing.Currency,
	}, nil
}

func handleScheduledClassEdit(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)
	lockTeacher := role == auth.RoleTeacher
	isSuperuser := role == auth.RoleSuperuser

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
		data, err := editScheduleData(ctx, scheduleID, lockTeacher, isSuperuser)
		if err != nil {
			HttpError(w, err.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		frontend.EditSchedule(data).Render(ctx, w)
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
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	currency := r.FormValue("currency")
	if err := validateScheduledClassRateCurrency(rate, currency); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	startTime := ""
	if existing.StartTime.Valid {
		startTime = existing.StartTime.String
	}

	if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
		ScheduleID:      scheduleID,
		StudentID:       studentID,
		TeacherID:       existing.TeacherID,
		Date:            existing.ScheduledDate,
		StartTime:       startTime,
		DurationMinutes: existing.DurationMinutes,
	}); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	err = dbRW.GetQueries().UpdateScheduledClassDetails(ctx, queries.UpdateScheduledClassDetailsParams{
		StudentID: studentID,
		Rate:      rate,
		Currency:  currency,
		ID:        scheduleID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if studentID != existing.StudentID && meetingSvc != nil {
		student, studentErr := dbRO.GetQueries().GetStudentByID(ctx, studentID)
		studentName := "student"
		if studentErr == nil {
			studentName = student.Name
		}
		if err := meetingSvc.SyncRoomForSchedule(ctx, meetings.ScheduledClassMeetingInput{
			ScheduleID:      scheduleID,
			TeacherID:       existing.TeacherID,
			StudentName:     studentName,
			ScheduledDate:   existing.ScheduledDate,
			StartTime:       startTime,
			DurationMinutes: existing.DurationMinutes,
		}); err != nil {
			logs.Log().Warn("zoom room sync failed after schedule edit",
				zap.Error(err),
				zap.Int64("schedule_id", scheduleID),
				zap.Int64("teacher_id", existing.TeacherID),
			)
		}
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "schedule", fmt.Sprintf("updated scheduled class id %d (student id %d, rate %.2f %s)", scheduleID, studentID, rate, currency))
	actor := auth.GetUser(ctx)
	notifyCrossParty(ctx, actor, existing.TeacherID, teacherNameByID(ctx, existing.TeacherID), notifications.KindScheduleChanged,
		fmt.Sprintf("Scheduled class on %s was updated", existing.ScheduledDate))

	if _, err := fmt.Fprint(w, "Scheduled class updated successfully!\n"); err != nil {
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
	if role == auth.RoleSuperuser {
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
	}
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
	if req.Rate <= 0 {
		return errors.New("rate must be greater than zero")
	}
	if req.Currency == "" {
		return errors.New("currency is required")
	}
	if !constants.ValidCurrency(req.Currency) {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}
	return nil
}

func scheduledClassTeacherAvatar(sc queries.GetScheduledClassesFilteredRow) models.AvatarView {
	hasPicture := sc.TeacherProfilePicture.Valid && sc.TeacherProfilePicture.String != ""
	assignedColor := sc.TeacherAssignedColor
	if assignedColor == "" {
		assignedColor = "#B9D283"
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
	}
}
