package cmd

import (
	"context"
	"database/sql"
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
	"zion-english/internal/notifications"
	"zion-english/internal/utils"
)

func handleClassesPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "classes", "/edit"); ok {
		handleClassEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "classes", "/view"); ok {
		handleClassView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "classes", "/delete"); ok {
		handleDeleteClassRecord(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func validateDeletionReason(reason string) error {
	if utils.IsBlank(reason) {
		return errors.New("reason is required for deletion")
	}
	return nil
}

func validateClassReason(reason string) error {
	if utils.IsBlank(reason) {
		return errors.New("reason is required")
	}
	return nil
}

func handleDeleteClassRecord(w http.ResponseWriter, r *http.Request, recordID int64) {
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

	err = dbRW.GetQueries().SoftDeleteClassRecord(ctx, queries.SoftDeleteClassRecordParams{
		Reason: sql.NullString{String: reason, Valid: true},
		ID:     recordID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLogAs(ctx, user, "classes", fmt.Sprintf("deleted class record id %d (student id %d, date %s, reason: %s)", recordID, existing.StudentID, existing.Date, reason))
	respondScheduledClassAction(w, r.FormValue("from"), "Class deleted successfully.")
}

func classEditClassData(ctx context.Context, recordID int64, readonly bool) (frontend.EditClassData, error) {
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	existing, err := dbRO.GetQueries().GetClassRecordByID(ctx, recordID)
	if err != nil {
		return frontend.EditClassData{}, err
	}

	rules := classrules.ClassRecordRules{DB: dbRO.GetQueries()}
	if err := rules.ValidateEditAccess(existing.TeacherID, user); err != nil {
		return frontend.EditClassData{}, err
	}

	startTime, endTime := classRecordTimesForEdit(existing.StartTime, existing.EndTime, existing.DurationMinutes)
	teacherRoles, err := loadTeacherRoles(ctx, existing.TeacherID)
	if err != nil {
		return frontend.EditClassData{}, err
	}
	materials, err := loadClassRecordLearningMaterialLinks(ctx, recordID)
	if err != nil {
		return frontend.EditClassData{}, err
	}
	return frontend.EditClassData{
		RecordID:        strconv.FormatInt(recordID, 10),
		Readonly:        readonly,
		IsSuperuser:     auth.HasAdminAccess(role),
		StudentID:       strconv.FormatInt(existing.StudentID, 10),
		TeacherID:       strconv.FormatInt(existing.TeacherID, 10),
		StudentName:     existing.StudentName,
		TeacherName:     existing.TeacherName,
		TeacherAvatar:   avatarWithTeacherRoles(classRecordTeacherAvatar(existing), teacherRoles),
		Date:            existing.Date,
		StartTime:       startTime,
		EndTime:         endTime,
		DurationMinutes: existing.DurationMinutes,
		Rate:            existing.Rate,
		Currency:        existing.Currency,
		Status:          constants.ClassListFilterStatus(existing.Status),
		Reason:          existing.Reason.String,
		Notes:           existing.Notes.String,
		LearningMaterials: materials,
	}, nil
}

func handleClassView(w http.ResponseWriter, r *http.Request, recordID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if _, err := dbRO.GetQueries().GetClassRecordByID(ctx, recordID); err != nil {
		HttpError(w, "Class record not found", http.StatusNotFound)
		return
	}

	data, err := classEditClassData(ctx, recordID, true)
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.ClassViewModal(data).Render(ctx, w)
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
		data, err := classEditClassData(ctx, recordID, false)
		if err != nil {
			HttpError(w, "Class record not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		frontend.EditClass(data).Render(ctx, w)
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
	duration, err := classRecordDurationFromForm(r)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	startTime, endTime := r.FormValue("start_time"), r.FormValue("end_time")
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	teacherID := existing.TeacherID
	if auth.HasAdminAccess(role) {
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
		StartTime:       startTime,
		EndTime:         endTime,
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
		StartTime:       sql.NullString{String: req.StartTime, Valid: req.StartTime != ""},
		EndTime:         sql.NullString{String: req.EndTime, Valid: req.EndTime != ""},
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

	if err := saveClassRecordLearningMaterials(ctx, user, recordID, parseLearningMaterialIDs(r)); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	updated, err := dbRW.GetQueries().GetClassRecordByID(ctx, recordID)
	if err == nil {
		insertAuditLogAs(ctx, auth.GetUser(ctx), "classes", formatClassRecordAudit(existing, updated))
		actor := auth.GetUser(ctx)
		notifyCrossParty(ctx, actor, updated.TeacherID, teacherNameByID(ctx, updated.TeacherID), notifications.KindClassUpdated,
			fmt.Sprintf("Class record updated for student on %s (status %s)", updated.Date, updated.Status))
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
	duration, err := classRecordDurationFromForm(r)
	if err != nil {
		return models.ClassRecordRequest{}, err
	}
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		return models.ClassRecordRequest{}, err
	}

	teacherID := defaultTeacherID
	if auth.HasAdminAccess(role) {
		teacherID, err = requireInt64(r.FormValue("teacher"))
		if err != nil {
			return models.ClassRecordRequest{}, errors.New("teacher is required")
		}
	}

	req := models.ClassRecordRequest{
		StudentID:       studentID,
		TeacherID:       teacherID,
		Date:            r.FormValue("date"),
		StartTime:       r.FormValue("start_time"),
		EndTime:         r.FormValue("end_time"),
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

func classRecordTeacherAvatar(cr queries.GetClassRecordByIDRow) frontend.AvatarProps {
	return buildTeacherListAvatarProps(
		cr.TeacherID,
		cr.TeacherFirstName,
		cr.TeacherMiddleName,
		cr.TeacherLastName,
		cr.TeacherAssignedColor,
		cr.TeacherProfilePicture,
	)
}

func classRecordViewFromRow(cr queries.GetClassRecordsFilteredRow) models.ClassRecordView {
	return models.ClassRecordView{
		ID:              cr.ID,
		StudentID:       cr.StudentID,
		TeacherID:       cr.TeacherID,
		StudentName:     cr.StudentName,
		TeacherName:     cr.TeacherName,
		Date:            cr.Date,
		StartTime:       cr.StartTime.String,
		EndTime:         cr.EndTime.String,
		DurationMinutes: cr.DurationMinutes,
		Rate:            cr.Rate,
		Currency:        cr.Currency,
		Status:          cr.Status,
		Reason:          cr.Reason.String,
		Notes:           cr.Notes.String,
		CreatedAt:       cr.CreatedAt,
		Source:          "record",
	}
}

func classesListTeacherAvatar(row queries.GetClassesListFilteredRow) models.AvatarView {
	hasPicture := row.TeacherProfilePicture.Valid && row.TeacherProfilePicture.String != ""
	assignedColor := row.TeacherAssignedColor
	if assignedColor == "" {
		assignedColor = constants.DefaultTeacherAssignedColor
	}
	return models.AvatarView{
		Initials:      utils.PersonInitials(row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName, row.TeacherName),
		AssignedColor: assignedColor,
		HasPicture:    hasPicture,
		PictureURL:    teacherPictureURL(row.TeacherID, hasPicture),
		Alt:           row.TeacherName + " avatar",
	}
}

func classesListViewFromRow(row queries.GetClassesListFilteredRow) models.ClassRecordView {
	startTime := ""
	if row.StartTime.Valid {
		startTime = row.StartTime.String
	}
	endTime := ""
	if row.EndTime.Valid {
		endTime = row.EndTime.String
	} else if startTime != "" {
		endTime = utils.EndTimeFromStartAndDuration(startTime, row.DurationMinutes)
	}
	reason := ""
	if row.Reason.Valid {
		reason = row.Reason.String
	}
	notes := ""
	if row.Notes.Valid {
		notes = row.Notes.String
	}
	return models.ClassRecordView{
		ID:              row.ID,
		StudentID:       row.StudentID,
		TeacherID:       row.TeacherID,
		StudentName:     row.StudentName,
		TeacherName:     row.TeacherName,
		TeacherAvatar:   classesListTeacherAvatar(row),
		Date:            row.Date,
		StartTime:       startTime,
		EndTime:         endTime,
		DurationMinutes: row.DurationMinutes,
		Rate:            row.Rate,
		Currency:        row.Currency,
		Status:          row.Status,
		Reason:          reason,
		Notes:           notes,
		CreatedAt:       row.CreatedAt,
		Source:          row.Source,
	}
}

func classesListCountParams(teacherID int64, startDate, endDate, statusFilter, nameFilter string) queries.CountClassesListFilteredParams {
	return queries.CountClassesListFilteredParams{
		Column1:         teacherID,
		TeacherID:       teacherID,
		Date:            startDate,
		Date_2:          endDate,
		Column5:         statusFilter,
		Column6:         statusFilter,
		Column7:         statusFilter,
		Column8:         statusFilter,
		Column9:         statusFilter,
		Column10:        statusFilter,
		Status:          statusFilter,
		Column12:        nameFilter,
		Column13:        sql.NullString{String: nameFilter, Valid: nameFilter != ""},
		Column14:        teacherID,
		TeacherID_2:     teacherID,
		ScheduledDate:   startDate,
		ScheduledDate_2: endDate,
		Column18:        statusFilter,
		Column19:        statusFilter,
		Column20:        statusFilter,
		Column21:        statusFilter,
		Column22:        statusFilter,
		Column23:        nameFilter,
		Column24:        sql.NullString{String: nameFilter, Valid: nameFilter != ""},
	}
}

func classesListListParams(teacherID int64, startDate, endDate, statusFilter, nameFilter string, limit, offset int64) queries.GetClassesListFilteredParams {
	return queries.GetClassesListFilteredParams{
		Column1:         teacherID,
		TeacherID:       teacherID,
		Date:            startDate,
		Date_2:          endDate,
		Column5:         statusFilter,
		Column6:         statusFilter,
		Column7:         statusFilter,
		Column8:         statusFilter,
		Column9:         statusFilter,
		Column10:        statusFilter,
		Status:          statusFilter,
		Column12:        nameFilter,
		Column13:        sql.NullString{String: nameFilter, Valid: nameFilter != ""},
		Column14:        teacherID,
		TeacherID_2:     teacherID,
		ScheduledDate:   startDate,
		ScheduledDate_2: endDate,
		Column18:        statusFilter,
		Column19:        statusFilter,
		Column20:        statusFilter,
		Column21:        statusFilter,
		Column22:        statusFilter,
		Column23:        nameFilter,
		Column24:        sql.NullString{String: nameFilter, Valid: nameFilter != ""},
		Limit:           limit,
		Offset:          offset,
	}
}

func classRecordsCountParams(teacherID int64, startDate, endDate, statusFilter, nameFilter string) queries.CountClassRecordsFilteredParams {
	return queries.CountClassRecordsFilteredParams{
		Column1:   teacherID,
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
		Column5:   statusFilter,
		Status:    statusFilter,
		Column7:   nameFilter,
		Column8:   sql.NullString{String: nameFilter, Valid: true},
	}
}

func classRecordsListParams(teacherID int64, startDate, endDate, statusFilter, nameFilter string, limit, offset int64) queries.GetClassRecordsFilteredParams {
	return queries.GetClassRecordsFilteredParams{
		Column1:   teacherID,
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
		Column5:   statusFilter,
		Status:    statusFilter,
		Column7:   nameFilter,
		Column8:   sql.NullString{String: nameFilter, Valid: true},
		Limit:     limit,
		Offset:    offset,
	}
}

func totalRateParams(teacherID int64, startDate, endDate string) queries.GetTotalRateByTeacherAndDateRangeParams {
	return queries.GetTotalRateByTeacherAndDateRangeParams{
		Column1:   teacherID,
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	}
}

type classRecordsQuery struct {
	teacherID    int64
	startDate    string
	endDate      string
	statusFilter string
	nameFilter   string
	showAll      bool
}

func parseClassRecordsQuery(r *http.Request) (classRecordsQuery, error) {
	startDate, endDate, err := parseListDateRange(r)
	if err != nil {
		return classRecordsQuery{}, err
	}

	teacherIDStr := strings.TrimSpace(r.URL.Query().Get("teacherId"))
	if teacherIDStr == "" {
		teacherIDStr = strings.TrimSpace(r.URL.Query().Get("teacher"))
	}
	statusFilter := r.URL.Query().Get("status")
	nameFilter := r.URL.Query().Get("q")

	role := auth.GetRole(r.Context())
	var teacherID int64
	showAll := false
	if teacherIDStr == "" {
		if !auth.HasAdminAccess(role) {
			return classRecordsQuery{}, errors.New("missing required parameters")
		}
		showAll = true
	} else {
		parsedID, err := strconv.ParseInt(teacherIDStr, 10, 64)
		if err != nil {
			return classRecordsQuery{}, errors.New("invalid teacher ID")
		}
		teacherID = parsedID
		if role == auth.RoleTeacher {
			user := auth.GetUser(r.Context())
			if teacherID != user.ID {
				return classRecordsQuery{}, errors.New("forbidden")
			}
		}
	}

	return classRecordsQuery{
		teacherID:    teacherID,
		startDate:    startDate,
		endDate:      endDate,
		statusFilter: statusFilter,
		nameFilter:   nameFilter,
		showAll:      showAll,
	}, nil
}

func handleClassRecordsPartial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q, err := parseClassRecordsQuery(r)
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

	total, err := dbRO.GetQueries().CountClassesListFiltered(ctx, classesListCountParams(q.teacherID, q.startDate, q.endDate, q.statusFilter, q.nameFilter))
	if err != nil {
		HttpError(w, "Failed to count class records", http.StatusInternalServerError)
		return
	}
	page.Total = total

	records, err := dbRO.GetQueries().GetClassesListFiltered(ctx, classesListListParams(q.teacherID, q.startDate, q.endDate, q.statusFilter, q.nameFilter, int64(page.Size), int64(page.Offset())))
	if err != nil {
		HttpError(w, "Failed to fetch class records", http.StatusInternalServerError)
		return
	}

	views := make([]models.ClassRecordView, 0, len(records))
	teacherIDs := make([]int64, 0, len(records))
	for _, cr := range records {
		views = append(views, classesListViewFromRow(cr))
		teacherIDs = append(teacherIDs, cr.TeacherID)
	}
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		HttpError(w, "Failed to fetch teacher roles", http.StatusInternalServerError)
		return
	}
	enrichClassRecordViewsWithRoleBadges(views, rolesMap)
	rows := frontend.ClassRecordRowFromViews(views)

	colspan := 7
	if q.showAll {
		colspan = 8
	}
	pagination := frontend.BuildPaginationData(page.Number, page.Size, total)
	partialsURL := utils.URL("/classes/partials/rows")
	includeSelector := "#classesToolbar"

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.ClassRecordsPartial(rows, q.showAll, colspan, "No classes found for the selected criteria", pagination, partialsURL, includeSelector).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleClassRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	if r.Method == http.MethodGet {
		prefill := parseRecordClassPrefill(r)
		learningMaterials, err := recordClassLearningMaterials(ctx, prefill)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		frontend.RecordClass(frontend.RecordClassData{
			IsSuperuser:         auth.HasAdminAccess(role),
			Prefill:             prefill,
			LearningMaterials:   learningMaterials,
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

	recordID, err := dbRW.GetQueries().InsertClassRecord(ctx, queries.InsertClassRecordParams{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.Date,
		StartTime:       sql.NullString{String: req.StartTime, Valid: req.StartTime != ""},
		EndTime:         sql.NullString{String: req.EndTime, Valid: req.EndTime != ""},
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

	materialIDs := parseLearningMaterialIDs(r)
	if len(materialIDs) == 0 {
		if fromSchedule := r.FormValue("fromSchedule"); fromSchedule != "" {
			scheduleID, parseErr := strconv.ParseInt(fromSchedule, 10, 64)
			if parseErr == nil && scheduleID > 0 {
				if copyErr := copyScheduledClassLearningMaterials(ctx, dbRW.GetQueries(), scheduleID, recordID); copyErr != nil {
					sendErrorLog(w, copyErr.Error())
					return
				}
			}
		}
	} else if err := saveClassRecordLearningMaterials(ctx, user, recordID, materialIDs); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if fromSchedule := r.FormValue("fromSchedule"); fromSchedule != "" {
		scheduleID, err := strconv.ParseInt(fromSchedule, 10, 64)
		if err != nil || scheduleID <= 0 {
			sendErrorLog(w, "invalid schedule reference")
			return
		}
		if err := markScheduledClassConducted(ctx, scheduleID, req); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
		insertAuditLogAs(ctx, auth.GetUser(ctx), "schedule", fmt.Sprintf("marked scheduled class id %d as %s", scheduleID, req.Status))
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "classes", fmt.Sprintf("recorded class for student id %d (teacher id %d, date %s, status %s)", req.StudentID, req.TeacherID, req.Date, req.Status))
	actor := auth.GetUser(ctx)
	notifyCrossParty(ctx, actor, req.TeacherID, teacherNameByID(ctx, req.TeacherID), notifications.KindClassRecorded,
		fmt.Sprintf("Class recorded for student on %s (status %s)", req.Date, req.Status))

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
	if !constants.ValidCurrency(req.Currency) {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}
	if !constants.ValidClassStatus(req.Status) {
		return errors.New("invalid status")
	}
	if req.Status != "conducted" && utils.IsBlank(req.Reason) {
		return errors.New("reason is required for cancelled or rescheduled classes")
	}
	return nil
}

func recordClassLearningMaterials(ctx context.Context, prefill models.RecordClassPrefill) ([]frontend.ClassLearningMaterialLink, error) {
	if !prefill.HasPrefill || prefill.FromSchedule == "" {
		return nil, nil
	}
	scheduleID, err := strconv.ParseInt(prefill.FromSchedule, 10, 64)
	if err != nil || scheduleID <= 0 {
		return nil, nil
	}
	return loadScheduledClassLearningMaterialLinks(ctx, scheduleID)
}

func parseRecordClassPrefill(r *http.Request) models.RecordClassPrefill {
	q := r.URL.Query()
	fromSchedule := q.Get("fromSchedule")
	if fromSchedule == "" {
		return models.RecordClassPrefill{}
	}

	prefill := models.RecordClassPrefill{
		FromSchedule:    fromSchedule,
		StudentID:       q.Get("student"),
		TeacherID:       q.Get("teacher"),
		Date:            q.Get("date"),
		DurationMinutes: q.Get("duration"),
		Rate:            q.Get("rate"),
		Currency:        q.Get("currency"),
		Status:          constants.ClassStatus(q.Get("status")),
		HasPrefill:      true,
	}
	if prefill.Status == "" {
		prefill.Status = constants.ClassStatusConducted
	}

	scheduleID, err := strconv.ParseInt(fromSchedule, 10, 64)
	if err != nil || scheduleID <= 0 {
		return models.RecordClassPrefill{}
	}

	existing, err := dbRO.GetQueries().GetScheduledClassByID(r.Context(), scheduleID)
	if err != nil {
		return models.RecordClassPrefill{}
	}

	if role := auth.GetRole(r.Context()); role == auth.RoleTeacher {
		user := auth.GetUser(r.Context())
		if existing.TeacherID != user.ID {
			return models.RecordClassPrefill{}
		}
	}

	prefill.StudentName = existing.StudentName
	prefill.TeacherName = existing.TeacherName
	prefill.StudentID = strconv.FormatInt(existing.StudentID, 10)
	prefill.TeacherID = strconv.FormatInt(existing.TeacherID, 10)
	prefill.Date = existing.ScheduledDate
	prefill.DurationMinutes = strconv.FormatInt(existing.DurationMinutes, 10)
	prefill.Rate = strconv.FormatFloat(existing.Rate, 'f', -1, 64)
	prefill.Currency = existing.Currency
	if existing.StartTime.Valid {
		prefill.StartTime = existing.StartTime.String
		prefill.EndTime = utils.EndTimeFromStartAndDuration(prefill.StartTime, existing.DurationMinutes)
	}

	return prefill
}

func classRecordDurationFromForm(r *http.Request) (int64, error) {
	startTime := r.FormValue("start_time")
	endTime := r.FormValue("end_time")
	if startTime == "" || endTime == "" {
		return 0, errors.New("start and end times are required")
	}
	duration, err := utils.DurationMinutesFromRange(startTime, endTime)
	if err != nil {
		return 0, friendlyTimeRangeError(err)
	}
	return duration, nil
}

func friendlyTimeRangeError(err error) error {
	switch {
	case errors.Is(err, utils.ErrEndBeforeStart):
		return errors.New("end time must be after start time")
	case errors.Is(err, utils.ErrInvalidStartTime):
		return errors.New("invalid start time")
	case errors.Is(err, utils.ErrInvalidEndTime):
		return errors.New("invalid end time")
	case errors.Is(err, utils.ErrTimeRequired):
		return errors.New("start and end times are required")
	default:
		return err
	}
}

func classRecordDefaultStartTime(durationMinutes int64) string {
	if durationMinutes >= 12*60 {
		return "00:00"
	}
	return "09:00"
}

func classRecordTimesForEdit(startTime, endTime sql.NullString, durationMinutes int64) (string, string) {
	start := classRecordDefaultStartTime(durationMinutes)
	if startTime.Valid && strings.TrimSpace(startTime.String) != "" {
		start = startTime.String
	}
	end := utils.EndTimeFromStartAndDuration(start, durationMinutes)
	if endTime.Valid && strings.TrimSpace(endTime.String) != "" {
		end = endTime.String
	}
	return start, end
}
