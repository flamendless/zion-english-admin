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
	"zion-english/internal/models"
	"zion-english/internal/utils"
)

func handleSchedulePath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "schedule", "/cancel"); ok {
		handleCancelScheduledClass(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "schedule", "/reschedule"); ok {
		handleRescheduleScheduledClass(w, r, id)
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
		DurationMinutes: req.DurationMinutes,
	}); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	var startTime sql.NullString
	if req.StartTime != "" {
		startTime = sql.NullString{String: req.StartTime, Valid: true}
	}

	err = dbRW.GetQueries().InsertScheduledClass(ctx, queries.InsertScheduledClassParams{
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

	insertAuditLog(ctx, "schedule", fmt.Sprintf("scheduled class for student id %d (teacher id %d, date %s)", req.StudentID, req.TeacherID, req.ScheduledDate))

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

func handleGetScheduledClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teacherIDStr := r.URL.Query().Get("teacherId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	statusFilter := r.URL.Query().Get("status")
	nameFilter := r.URL.Query().Get("q")

	if startDate == "" || endDate == "" {
		HttpError(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	role := auth.GetRole(r.Context())
	var teacherID int64
	if teacherIDStr == "" {
		if role != auth.RoleSuperuser {
			HttpError(w, "Missing required parameters", http.StatusBadRequest)
			return
		}
	} else {
		parsedID, err := strconv.ParseInt(teacherIDStr, 10, 64)
		if err != nil {
			HttpError(w, "Invalid teacher ID", http.StatusBadRequest)
			return
		}
		teacherID = parsedID
		if role == auth.RoleTeacher {
			user := auth.GetUser(r.Context())
			if teacherID != user.ID {
				HttpError(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
	}

	page := utils.ParsePageQuery(r)
	ctx := r.Context()

	countParams := scheduledClassesCountParams(teacherID, startDate, endDate, statusFilter, nameFilter)
	total, err := dbRO.GetQueries().CountScheduledClassesFiltered(ctx, countParams)
	if err != nil {
		HttpError(w, "Failed to count scheduled classes", http.StatusInternalServerError)
		return
	}
	page.Total = total

	records, err := dbRO.GetQueries().GetScheduledClassesFiltered(ctx, scheduledClassesListParams(teacherID, startDate, endDate, statusFilter, nameFilter, int64(page.Size), int64(page.Offset())))
	if err != nil {
		HttpError(w, "Failed to fetch scheduled classes", http.StatusInternalServerError)
		return
	}

	var response []models.ScheduledClassView
	for _, sc := range records {
		response = append(response, scheduledClassViewFromRow(sc))
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
	reason := r.FormValue("reason")

	err = dbRW.GetQueries().UpdateScheduledClassStatus(ctx, queries.UpdateScheduledClassStatusParams{
		Status:   "cancelled",
		Reason:   sql.NullString{String: reason, Valid: reason != ""},
		ID:       scheduleID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLog(ctx, "schedule", fmt.Sprintf("cancelled scheduled class id %d (student id %d, date %s)", scheduleID, existing.StudentID, existing.ScheduledDate))

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
	reason := r.FormValue("reason")

	if err := rules.Validate(ctx, user, classrules.ScheduledClassInput{
		ScheduleID:      scheduleID,
		StudentID:       existing.StudentID,
		TeacherID:       existing.TeacherID,
		Date:            newDate,
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
		Reason:        sql.NullString{String: reason, Valid: reason != ""},
		ID:            scheduleID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLog(ctx, "schedule", fmt.Sprintf("rescheduled class id %d from %s to %s", scheduleID, existing.ScheduledDate, newDate))

	if _, err := fmt.Fprint(w, "Class rescheduled.\n"); err != nil {
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
