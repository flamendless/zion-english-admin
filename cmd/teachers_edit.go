package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/notifications"
	"zion-english/internal/processor"
	"zion-english/internal/teachers"
	"zion-english/internal/utils"
)

func teacherFilterParams(q, status string) queries.CountTeachersFilteredParams {
	qNull := sql.NullString{String: q, Valid: q != ""}
	return queries.CountTeachersFilteredParams{
		Column1: q,
		Column2: qNull,
		Column3: qNull,
		Column4: status,
		Column5: status,
		Column6: status,
		Column7: status,
		Status:  status,
	}
}

func loadTeacherRoles(ctx context.Context, teacherID int64) ([]constants.TeacherRole, error) {
	rows, err := dbRO.GetQueries().GetTeacherRolesByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	return teachers.StringsToRoles(rows), nil
}

func teacherHasAdminRole(roles []constants.TeacherRole) bool {
	return slices.Contains(roles, constants.TeacherRoleAdmin)
}

func availableRoleOptions(assigned []constants.TeacherRole) []constants.TeacherRole {
	assignedSet := make(map[constants.TeacherRole]struct{}, len(assigned))
	for _, role := range assigned {
		assignedSet[role] = struct{}{}
	}
	options := make([]constants.TeacherRole, 0, len(constants.AllTeacherRoles()))
	for _, role := range constants.AllTeacherRoles() {
		if _, ok := assignedSet[role]; !ok {
			options = append(options, role)
		}
	}
	return options
}

func syncTeacherRoles(ctx context.Context, q *queries.Queries, teacherID int64, roles []constants.TeacherRole) error {
	if err := q.DeleteTeacherRolesByTeacherID(ctx, teacherID); err != nil {
		return err
	}
	for _, role := range roles {
		if err := q.InsertTeacherRole(ctx, queries.InsertTeacherRoleParams{
			TeacherID: teacherID,
			Role:      string(role),
		}); err != nil {
			return err
		}
	}
	return nil
}

func handleTeachersPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "teachers", "/edit"); ok {
		handleTeacherEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "teachers", "/view"); ok {
		handleTeacherView(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleTeacherView(w http.ResponseWriter, r *http.Request, teacherID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		HttpError(w, "Teacher not found", http.StatusNotFound)
		return
	}

	certifications := ""
	if row.Certifications.Valid {
		certifications = row.Certifications.String
	}
	sex := ""
	if row.Sex.Valid {
		sex = row.Sex.String
	}

	roleStrings, err := loadTeacherRoles(ctx, teacherID)
	if err != nil {
		HttpError(w, "Failed to load teacher roles", http.StatusInternalServerError)
		return
	}

	teacherName := utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName)
	zoomConnected, zoomConfigured, zoomConnectionsAllowed := profileZoomStatus(ctx, teacherID)
	googleCalendarConnected, googleCalendarConfigured, googleCalendarConnectionsAllowed := profileGoogleCalendarStatus(ctx, teacherID)
	w.Header().Set("Content-Type", "text/html")
	frontend.TeacherViewModal(frontend.TeacherViewData{
		ID:             strconv.FormatInt(teacherID, 10),
		Name:           teacherName,
		Email:          row.Email,
		FirstName:      row.FirstName,
		MiddleName:     row.MiddleName,
		LastName:       row.LastName,
		Birthdate:      row.Birthdate,
		Address:        row.Address,
		JoiningDate:    row.JoiningDate,
		MobileNumber:   row.MobileNumber,
		Certifications: certifications,
		AssignedColor:  row.AssignedColor,
		RatePerClass:   row.RatePerClass,
		Currency:       row.Currency,
		DriveUrl:       row.DriveUrl,
		Sex:            sex,
		Status:         constants.TeacherStatus(row.Status),
		Roles:          roleStrings,
		Avatar:         avatarWithTeacherRoles(buildTeacherAvatarProps(row), roleStrings),
		ZoomConfigured:                   zoomConfigured,
		ZoomConnected:                    zoomConnected,
		ZoomConnectionsAllowed:           zoomConnectionsAllowed,
		GoogleCalendarConfigured:         googleCalendarConfigured,
		GoogleCalendarConnected:          googleCalendarConnected,
		GoogleCalendarConnectionsAllowed: googleCalendarConnectionsAllowed,
	}).Render(ctx, w)
}

func handleTeachers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	sort := parseListSort(r, frontend.ListSortKindTeacher)
	page := utils.ParsePageQuery(r)

	filter := teacherFilterParams(q, status)
	total, err := dbRO.GetQueries().CountTeachersFiltered(ctx, filter)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count teachers: %v", err), http.StatusInternalServerError)
		return
	}
	page.Total = total

	allTeachers, err := dbRO.GetQueries().GetTeachersFiltered(ctx, queries.GetTeachersFilteredParams{
		Column1: filter.Column1,
		Column2: filter.Column2,
		Column3: filter.Column3,
		Column4: filter.Column4,
		Column5: filter.Column5,
		Column6: filter.Column6,
		Column7: filter.Column7,
		Status:  filter.Status,
		Limit:   total,
		Offset:  0,
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch teachers: %v", err), http.StatusInternalServerError)
		return
	}
	sortTeacherRows(allTeachers, sort)
	teachers := paginateSlice(allTeachers, page)

	teacherIDs := make([]int64, len(teachers))
	for i, t := range teachers {
		teacherIDs[i] = t.ID
	}

	docsStatusByTeacher := make(map[int64]string)
	rolesByTeacher := make(map[int64][]constants.TeacherRole)
	if len(teacherIDs) > 0 {
		docRows, err := dbRO.GetQueries().GetLatestTeacherDocumentStatusesByTeacherIDs(ctx, teacherIDs)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to fetch teacher document status: %v", err), http.StatusInternalServerError)
			return
		}
		for _, row := range docRows {
			if _, ok := docsStatusByTeacher[row.TeacherID]; !ok {
				docsStatusByTeacher[row.TeacherID] = row.Status
			}
		}

		roleRows, err := dbRO.GetQueries().GetTeacherRolesByTeacherIDs(ctx, teacherIDs)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to fetch teacher roles: %v", err), http.StatusInternalServerError)
			return
		}
		for _, row := range roleRows {
			rolesByTeacher[row.TeacherID] = append(rolesByTeacher[row.TeacherID], constants.TeacherRole(row.Role))
		}
	}

	viewTeachers := make([]frontend.TeacherItem, len(teachers))
	for i, t := range teachers {
		teacherName := utils.ComposePersonName(t.FirstName, t.MiddleName, t.LastName)
		viewTeachers[i] = frontend.TeacherItem{
			ID:             strconv.FormatInt(t.ID, 10),
			Name:           teacherName,
			Birthdate:      t.Birthdate,
			Address:        t.Address,
			JoiningDate:    t.JoiningDate,
			MobileNumber:   t.MobileNumber,
			Email:          t.Email,
			Certifications: t.Certifications.String,
			AssignedColor:  t.AssignedColor,
			RatePerClass:   t.RatePerClass,
			Currency:       t.Currency,
			DriveUrl:       t.DriveUrl,
			Sex:            t.Sex.String,
			Status:         constants.TeacherStatus(t.Status),
			DocsStatus:     constants.TeacherDocumentStatus(docsStatusByTeacher[t.ID]),
			Roles:          rolesByTeacher[t.ID],
			Deleted:        t.Deleted != 0,
			CreatedAt:      utils.FormatNullDateTimeSecondsPHT(t.CreatedAt),
			Avatar: avatarWithTeacherRoles(buildTeacherListAvatarProps(
				t.ID, t.FirstName, t.MiddleName, t.LastName, t.AssignedColor, t.ProfilePicture,
			), rolesByTeacher[t.ID]),
		}
	}

	params := listQueryParamsWithSort(r, frontend.ListSortKindTeacher)
	w.Header().Set("Content-Type", "text/html")
	frontend.Teachers(frontend.TeacherData{
		Teachers:       viewTeachers,
		Query:          q,
		Status:         constants.TeacherFilterStatus(status),
		SortBy:         sort.By,
		SortOrder:      string(sort.Order),
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(utils.URL("/teachers"), page.Number-1, page.Size, params),
		NextURL:        utils.BuildPageURLAt(utils.URL("/teachers"), page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     utils.URL("/teachers"),
	}).Render(ctx, w)
}

func handleTeacherEdit(w http.ResponseWriter, r *http.Request, teacherID int64) {
	ctx := r.Context()
	actorRole := auth.GetRole(ctx)

	existing, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		HttpError(w, "Teacher not found", http.StatusNotFound)
		return
	}
	if existing.Deleted != 0 {
		HttpError(w, "Teacher not found", http.StatusNotFound)
		return
	}

	existingRoles, err := loadTeacherRoles(ctx, teacherID)
	if err != nil {
		HttpError(w, "Failed to load teacher roles", http.StatusInternalServerError)
		return
	}
	targetHasAdmin := teacherHasAdminRole(existingRoles)
	canManageRoles := teachers.CanManageTeacherRoles(string(actorRole), targetHasAdmin)

	if r.Method == http.MethodGet {
		template := ""
		if existing.Template.Valid {
			template = existing.Template.String
		}
		w.Header().Set("Content-Type", "text/html")
		frontend.EditTeacher(frontend.EditTeacherData{
			ID:             strconv.FormatInt(teacherID, 10),
			FirstName:      existing.FirstName,
			MiddleName:     existing.MiddleName,
			LastName:       existing.LastName,
			Birthdate:      existing.Birthdate,
			Address:        existing.Address,
			JoiningDate:    existing.JoiningDate,
			MobileNumber:   existing.MobileNumber,
			Email:          existing.Email,
			Certifications: existing.Certifications.String,
			AssignedColor:  existing.AssignedColor,
			RatePerClass:   existing.RatePerClass,
			Currency:       existing.Currency,
			DriveUrl:       existing.DriveUrl,
			Sex:            existing.Sex.String,
			Template:       template,
			CanManageRoles: canManageRoles,
			Roles:          existingRoles,
			RoleOptions:    availableRoleOptions(existingRoles),
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

	req, err := parseTeacherForm(r, false)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if err := validateTeacherFields(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	template := strings.TrimSpace(r.FormValue("template"))
	if err := processor.ValidateSheetTemplate(template); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	mobileCount, err := dbRO.GetQueries().GetTeacherCountByMobileExcludingID(ctx, queries.GetTeacherCountByMobileExcludingIDParams{
		MobileNumber: req.MobileNumber,
		ID:           teacherID,
	})
	if err != nil || mobileCount > 0 {
		sendErrorLog(w, "a teacher with this mobile number already exists")
		return
	}

	emailCount, err := dbRO.GetQueries().GetTeacherCountByEmailExcludingID(ctx, queries.GetTeacherCountByEmailExcludingIDParams{
		Email: req.Email,
		ID:    teacherID,
	})
	if err != nil || emailCount > 0 {
		sendErrorLog(w, "a teacher with this email already exists")
		return
	}

	var submittedRoles []constants.TeacherRole
	if canManageRoles {
		submittedRoles, err = teachers.ParseTeacherRoles(r.Form["roles"])
		if err != nil {
			sendErrorLog(w, err.Error())
			return
		}
		if err := teachers.ValidateRoleAssignment(string(actorRole), targetHasAdmin, submittedRoles); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	tx, err := dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		sendErrorLog(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	qtx := dbRW.GetQueries().WithTx(tx)

	err = qtx.UpdateTeacherBySuperuser(ctx, queries.UpdateTeacherBySuperuserParams{
		FirstName:      req.FirstName,
		MiddleName:     req.MiddleName,
		LastName:       req.LastName,
		Birthdate:      req.Birthdate,
		Address:        req.Address,
		JoiningDate:    req.JoiningDate,
		MobileNumber:   req.MobileNumber,
		Email:          req.Email,
		Certifications: sql.NullString{String: req.Certifications, Valid: req.Certifications != ""},
		AssignedColor:  req.AssignedColor,
		RatePerClass:   req.RatePerClass,
		Currency:       req.Currency,
		DriveUrl:       req.DriveUrl,
		Sex:            sql.NullString{String: req.Sex, Valid: req.Sex != ""},
		Template:       sql.NullString{String: template, Valid: template != ""},
		ID:             teacherID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if canManageRoles {
		if err := syncTeacherRoles(ctx, qtx, teacherID, submittedRoles); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	updated, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	auditMessage := formatTeacherAudit(existing, updated)
	if canManageRoles {
		if roleDiff := teachers.FormatRoleDiff(existingRoles, submittedRoles); roleDiff != "" {
			auditMessage += "; roles: " + roleDiff
		}
	}
	insertAuditLogAs(ctx, auth.GetUser(ctx), "teachers", auditMessage)
	teacherName := utils.ComposePersonName(updated.FirstName, updated.MiddleName, updated.LastName)
	notifyTeacher(ctx, teacherID, teacherName, auth.GetUser(ctx), notifications.KindProfileUpdated,
		"Your profile was updated by an administrator", "")

	if _, err := fmt.Fprint(w, "Teacher updated successfully!\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
}
