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
	"zion-english/internal/database/queries"
	"zion-english/internal/models"
	"zion-english/internal/utils"
)

func handleClassesPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "classes", "/edit"); ok {
		handleClassEdit(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleClassEdit(w http.ResponseWriter, r *http.Request, recordID int64) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	existing, err := dbRO.GetQueries().GetClassRecordByID(ctx, recordID)
	if err != nil {
		HttpError(w, "Class record not found", http.StatusNotFound)
		return
	}

	rules := classrules.ClassRecordRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateEditAccess(existing.TeacherID, user); err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.EditClass(frontend.EditClassData{
			RecordID:        strconv.FormatInt(recordID, 10),
			IsSuperuser:     role == auth.RoleSuperuser,
			StudentID:       strconv.FormatInt(existing.StudentID, 10),
			TeacherID:       strconv.FormatInt(existing.TeacherID, 10),
			StudentName:     existing.StudentName,
			TeacherName:     existing.TeacherName,
			Date:            existing.Date,
			DurationMinutes: existing.DurationMinutes,
			Rate:            existing.Rate,
			Currency:        existing.Currency,
			Status:          existing.Status,
			Reason:          existing.Reason.String,
			Notes:           existing.Notes.String,
		}).Render(ctx, w)
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

	studentID, err := requireInt64(r.FormValue("student"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	duration, err := requireInt64(r.FormValue("duration"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	teacherID := existing.TeacherID
	if role == auth.RoleSuperuser {
		teacherID, err = requireInt64(r.FormValue("teacher"))
		if err != nil {
			sendErrorLog(w, "teacher is required")
			return
		}
	}

	req := models.ClassRecordRequest{
		StudentID:       studentID,
		TeacherID:       teacherID,
		Date:            r.FormValue("date"),
		DurationMinutes: duration,
		Rate:            rate,
		Currency:        r.FormValue("currency"),
		Status:          r.FormValue("status"),
		Reason:          r.FormValue("reason"),
		Notes:           r.FormValue("notes"),
	}

	if err := validateClassRecordRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if err := rules.Validate(ctx, user, classrules.ClassRecordInput{
		RecordID:        recordID,
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.Date,
		DurationMinutes: req.DurationMinutes,
	}); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	err = dbRW.GetQueries().UpdateClassRecord(ctx, queries.UpdateClassRecordParams{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.Date,
		DurationMinutes: req.DurationMinutes,
		Rate:            req.Rate,
		Currency:        req.Currency,
		Status:          req.Status,
		Reason:          sql.NullString{String: req.Reason, Valid: req.Reason != ""},
		Notes:           sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		ID:              recordID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	updated, err := dbRW.GetQueries().GetClassRecordByID(ctx, recordID)
	if err == nil {
		insertAuditLog(ctx, "classes", formatClassRecordAudit(existing, updated))
	}

	if _, err := fmt.Fprint(w, "Class updated successfully!\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
}

func parseClassRecordRequest(r *http.Request, user auth.User, role auth.Role, defaultTeacherID int64) (models.ClassRecordRequest, error) {
	studentID, err := requireInt64(r.FormValue("student"))
	if err != nil {
		return models.ClassRecordRequest{}, err
	}
	duration, err := requireInt64(r.FormValue("duration"))
	if err != nil {
		return models.ClassRecordRequest{}, err
	}
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		return models.ClassRecordRequest{}, err
	}

	teacherID := defaultTeacherID
	if role == auth.RoleSuperuser {
		teacherID, err = requireInt64(r.FormValue("teacher"))
		if err != nil {
			return models.ClassRecordRequest{}, errors.New("teacher is required")
		}
	}

	req := models.ClassRecordRequest{
		StudentID:       studentID,
		TeacherID:       teacherID,
		Date:            r.FormValue("date"),
		DurationMinutes: duration,
		Rate:            rate,
		Currency:        r.FormValue("currency"),
		Status:          r.FormValue("status"),
		Reason:          r.FormValue("reason"),
		Notes:           r.FormValue("notes"),
	}
	return req, validateClassRecordRequest(&req)
}

func applyClassRecordRules(ctx context.Context, user auth.User, req models.ClassRecordRequest, recordID int64) error {
	return classrules.ClassRecordRules{DB: dbRO.GetQueries()}.Validate(ctx, user, classrules.ClassRecordInput{
		RecordID:        recordID,
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.Date,
		DurationMinutes: req.DurationMinutes,
	})
}

func classRecordViewFromRow(cr queries.GetClassRecordsFilteredRow) models.ClassRecordView {
	return models.ClassRecordView{
		ID:              cr.ID,
		StudentID:       cr.StudentID,
		TeacherID:       cr.TeacherID,
		StudentName:     cr.StudentName,
		TeacherName:     cr.TeacherName,
		Date:            cr.Date,
		DurationMinutes: cr.DurationMinutes,
		Rate:            cr.Rate,
		Currency:        cr.Currency,
		Status:          cr.Status,
		Reason:          cr.Reason.String,
		Notes:           cr.Notes.String,
		CreatedAt:       cr.CreatedAt,
	}
}

func handleGetClassRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teacherIDStr := r.URL.Query().Get("teacherId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	statusFilter := r.URL.Query().Get("status")
	nameFilter := r.URL.Query().Get("q")

	if teacherIDStr == "" || startDate == "" || endDate == "" {
		HttpError(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64)
	if err != nil {
		HttpError(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	if auth.GetRole(r.Context()) == auth.RoleTeacher {
		user := auth.GetUser(r.Context())
		if teacherID != user.ID {
			HttpError(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	page := utils.ParsePageQuery(r)
	ctx := r.Context()

		total, err := dbRO.GetQueries().CountClassRecordsFiltered(ctx, queries.CountClassRecordsFilteredParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
		Column4:   statusFilter,
		Status:    statusFilter,
		Column6:   nameFilter,
		Column7:   sql.NullString{String: nameFilter, Valid: true},
	})
	if err != nil {
		HttpError(w, "Failed to count class records", http.StatusInternalServerError)
		return
	}
	page.Total = total

	records, err := dbRO.GetQueries().GetClassRecordsFiltered(ctx, queries.GetClassRecordsFilteredParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
		Column4:   statusFilter,
		Status:    statusFilter,
		Column6:   nameFilter,
		Column7:   sql.NullString{String: nameFilter, Valid: true},
		Limit:     int64(page.Size),
		Offset:    int64(page.Offset()),
	})
	if err != nil {
		HttpError(w, "Failed to fetch class records", http.StatusInternalServerError)
		return
	}

	totalRate, err := dbRO.GetQueries().GetTotalRateByTeacherAndDateRange(ctx, queries.GetTotalRateByTeacherAndDateRangeParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		HttpError(w, "Failed to fetch total rate", http.StatusInternalServerError)
		return
	}

	var response []models.ClassRecordView
	for _, cr := range records {
		response = append(response, classRecordViewFromRow(cr))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records":   response,
		"totalRate": totalRate,
		"page": map[string]interface{}{
			"number":     page.Number,
			"size":       page.Size,
			"total":      page.Total,
			"totalPages": page.TotalPages(),
		},
	})
}

func handleClassRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.RecordClass(frontend.RecordClassData{
			IsSuperuser: role == auth.RoleSuperuser,
		}).Render(ctx, w)
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if role == auth.RoleTeacher && user.ID == 0 {
		HttpError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rl := &requestLogs{}
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	req, err := parseClassRecordRequest(r, user, role, user.ID)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if err := applyClassRecordRules(ctx, user, req, 0); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	err = dbRW.GetQueries().InsertClassRecord(ctx, queries.InsertClassRecordParams{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.Date,
		DurationMinutes: req.DurationMinutes,
		Rate:            req.Rate,
		Currency:        req.Currency,
		Status:          req.Status,
		Reason:          sql.NullString{String: req.Reason, Valid: req.Reason != ""},
		Notes:           sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		RecordedByRole:  string(user.Role),
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLog(ctx, "classes", fmt.Sprintf("recorded class for student id %d (teacher id %d, date %s, status %s)", req.StudentID, req.TeacherID, req.Date, req.Status))

	if _, err := fmt.Fprint(w, "Class recorded successfully!\n"); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rl.add("Class recorded successfully")
	for _, log := range rl.messages {
		if _, err := fmt.Fprint(w, log+"\n"); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}
}

func handleClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := frontend.ClassesData{}
		if auth.GetRole(r.Context()) == auth.RoleTeacher {
			user := auth.GetUser(r.Context())
			data.LockTeacher = true
			data.TeacherID = strconv.FormatInt(user.ID, 10)
			data.TeacherName = user.Name
		}
		w.Header().Set("Content-Type", "text/html")
		frontend.Classes(data).Render(r.Context(), w)
		return
	}

	HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func validateClassRecordRequest(req *models.ClassRecordRequest) error {
	if req.StudentID == 0 {
		return errors.New("student is required")
	}
	if req.TeacherID == 0 {
		return errors.New("teacher is required")
	}
	if req.Date == "" {
		return errors.New("date is required")
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
	validCurrencies := map[string]bool{"KRW": true, "CAD": true, "YEN": true, "PHP": true}
	if !validCurrencies[req.Currency] {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}
	validStatuses := map[string]bool{"conducted": true, "cancelled": true, "rescheduled": true}
	if !validStatuses[req.Status] {
		return errors.New("invalid status")
	}
	if req.Status != "conducted" && strings.TrimSpace(req.Reason) == "" {
		return errors.New("reason is required for cancelled or rescheduled classes")
	}
	return nil
}
