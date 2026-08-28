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
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/models"
	"zion-english/internal/utils"
)

func parseAssignedTeacherIDs(ctx context.Context, q *queries.Queries, raw []string) ([]int64, error) {
	if len(raw) == 0 {
		return nil, errors.New("at least one assigned teacher is required")
	}
	seen := make(map[int64]bool)
	var ids []int64
	for _, tidStr := range raw {
		tid, err := strconv.ParseInt(tidStr, 10, 64)
		if err != nil || tid == 0 {
			continue
		}
		if seen[tid] {
			return nil, errors.New("duplicate teacher assignment")
		}
		seen[tid] = true
		teacher, err := q.GetTeacherByID(ctx, tid)
		if err != nil || teacher.Status != "approved" {
			return nil, errors.New("invalid assigned teacher")
		}
		ids = append(ids, tid)
	}
	if len(ids) == 0 {
		return nil, errors.New("at least one assigned teacher is required")
	}
	return ids, nil
}

func requireStudentAssignedToTeacher(ctx context.Context, teacherID, studentID int64) error {
	assigned, err := dbRO.GetQueries().IsStudentAssignedToTeacher(ctx, queries.IsStudentAssignedToTeacherParams{
		TeacherID: teacherID,
		StudentID: studentID,
	})
	if err != nil {
		return err
	}
	if assigned == 0 {
		return errors.New("student is not assigned to this teacher")
	}
	return nil
}

func studentFilterParams(q, status string, teacherID int64) queries.CountStudentsFilteredParams {
	return queries.CountStudentsFilteredParams{
		Column1:   q,
		Column2:   sql.NullString{String: q, Valid: true},
		Column3:   status,
		Status:    status,
		Column5:   teacherID,
		TeacherID: teacherID,
	}
}

func handleStudentsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "students", "/edit"); ok {
		handleStudentEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "students", "/view"); ok {
		handleStudentView(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func studentEditStudentData(ctx context.Context, studentID int64, readonly bool) (frontend.EditStudentData, error) {
	existing, err := dbRO.GetQueries().GetStudentByID(ctx, studentID)
	if err != nil {
		return frontend.EditStudentData{}, err
	}

	assignedTeachers, err := dbRO.GetQueries().GetTeachersByStudentID(ctx, studentID)
	if err != nil {
		return frontend.EditStudentData{}, err
	}

	teacherIDsForRoles := make([]int64, len(assignedTeachers))
	for i, t := range assignedTeachers {
		teacherIDsForRoles[i] = t.ID
	}
	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDsForRoles))
	if err != nil {
		return frontend.EditStudentData{}, err
	}

	teacherIDs := make([]string, len(assignedTeachers))
	teacherNames := make([]string, len(assignedTeachers))
	teachers := make([]frontend.TeacherListItem, len(assignedTeachers))
	for i, t := range assignedTeachers {
		name := utils.ComposePersonName(t.FirstName, t.MiddleName, t.LastName)
		teacherIDs[i] = strconv.FormatInt(t.ID, 10)
		teacherNames[i] = name
		teachers[i] = frontend.TeacherListItem{
			Name: name,
			Avatar: avatarWithTeacherRoles(
				buildTeacherListAvatarProps(t.ID, t.FirstName, t.MiddleName, t.LastName, t.AssignedColor, t.ProfilePicture),
				rolesMap[t.ID],
			),
		}
	}

	relationships, err := dbRO.GetQueries().GetRelationshipsByStudentID(ctx, studentID)
	if err != nil {
		return frontend.EditStudentData{}, err
	}

	relItems := make([]frontend.StudentRelationshipItem, len(relationships))
	for i, rel := range relationships {
		relationship := ""
		if rel.Relationship.Valid {
			relationship = rel.Relationship.String
		}
		relItems[i] = frontend.StudentRelationshipItem{
			RelatedStudentID:   strconv.FormatInt(rel.RelatedStudentID, 10),
			RelatedStudentName: rel.RelatedStudentName,
			Relationship:       relationship,
		}
	}

	inactiveReason := ""
	if existing.InactiveReason.Valid {
		inactiveReason = existing.InactiveReason.String
	}

	return frontend.EditStudentData{
		ID:             strconv.FormatInt(studentID, 10),
		Readonly:       readonly,
		Name:           existing.Name,
		Currency:       existing.Currency,
		Contact:        existing.Contact.String,
		RatePerClass:   existing.RatePerClass,
		ParentName:     existing.ParentName.String,
		AssignedColor:  existing.AssignedColor,
		Status:         constants.StudentStatus(existing.Status),
		InactiveReason: inactiveReason,
		TeacherIDs:     teacherIDs,
		TeacherNames:   teacherNames,
		Teachers:       teachers,
		Relationships:  relItems,
	}, nil
}

func saveStudentRelationships(ctx context.Context, studentID int64, r *http.Request) error {
	for _, relatedIDStr := range r.Form["removeRelationship"] {
		relatedID, err := strconv.ParseInt(relatedIDStr, 10, 64)
		if err != nil || relatedID <= 0 {
			continue
		}
		if err := dbRW.GetQueries().DeleteStudentRelationship(ctx, queries.DeleteStudentRelationshipParams{
			StudentID:        studentID,
			RelatedStudentID: relatedID,
		}); err != nil {
			return err
		}
	}

	relatedStudentID := int64(0)
	if relatedStudentValue := strings.TrimSpace(r.FormValue("relatedStudentId")); relatedStudentValue != "" {
		var err error
		relatedStudentID, err = strconv.ParseInt(relatedStudentValue, 10, 64)
		if err != nil || relatedStudentID <= 0 {
			return errors.New("invalid related student")
		}
	}

	if relatedStudentID == 0 {
		return nil
	}

	if relatedStudentID == studentID {
		return errors.New("a student cannot be related to themselves")
	}

	if _, err := dbRO.GetQueries().GetStudentByID(ctx, relatedStudentID); err != nil {
		return errors.New("related student not found")
	}

	relationship := strings.TrimSpace(r.FormValue("relationship"))
	return dbRW.GetQueries().InsertStudentRelationship(ctx, queries.InsertStudentRelationshipParams{
		StudentID:        studentID,
		RelatedStudentID: relatedStudentID,
		Relationship:     sql.NullString{String: relationship, Valid: relationship != ""},
	})
}

func handleStudentView(w http.ResponseWriter, r *http.Request, studentID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !auth.HasAdminAccess(auth.GetRole(r.Context())) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	data, err := studentEditStudentData(ctx, studentID, true)
	if err != nil {
		HttpError(w, "Student not found", http.StatusNotFound)
		return
	}
	data.IsSuperuser = true

	w.Header().Set("Content-Type", "text/html")
	frontend.StudentViewModal(data).Render(ctx, w)
}

func handleStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	teacherID := utils.QueryParamInt64(r, "teacherId")
	page := utils.ParsePageQuery(r)

	filter := studentFilterParams(q, status, teacherID)
	total, err := dbRO.GetQueries().CountStudentsFiltered(ctx, filter)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count students: %v", err), http.StatusInternalServerError)
		return
	}
	page.Total = total

	students, err := dbRO.GetQueries().GetStudentsFiltered(ctx, queries.GetStudentsFilteredParams{
		Column1:   filter.Column1,
		Column2:   filter.Column2,
		Column3:   filter.Column3,
		Status:    filter.Status,
		Column5:   filter.Column5,
		TeacherID: filter.TeacherID,
		Limit:     int64(page.Size),
		Offset:    int64(page.Offset()),
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch students: %v", err), http.StatusInternalServerError)
		return
	}

	relationships, err := dbRO.GetQueries().GetAllStudentRelationships(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch student relationships: %v", err), http.StatusInternalServerError)
		return
	}

	relationshipsByStudent := make(map[int64][]queries.GetAllStudentRelationshipsRow)
	for _, rel := range relationships {
		relationshipsByStudent[rel.StudentID] = append(relationshipsByStudent[rel.StudentID], rel)
	}

	teacherAssignments, err := dbRO.GetQueries().GetAllStudentTeacherNames(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch teacher assignments: %v", err), http.StatusInternalServerError)
		return
	}

	teachersByStudent := make(map[int64][]string)
	for _, assignment := range teacherAssignments {
		teachersByStudent[assignment.StudentID] = append(teachersByStudent[assignment.StudentID], assignment.TeacherName)
	}

	viewStudents := make([]frontend.StudentItem, len(students))
	for i, s := range students {
		viewStudents[i] = frontend.StudentItem{
			ID:               strconv.FormatInt(s.ID, 10),
			Name:             s.Name,
			Currency:         s.Currency,
			Contact:          s.Contact.String,
			RatePerClass:     s.RatePerClass,
			ParentName:       s.ParentName.String,
			AssignedColor:    s.AssignedColor,
			Status:           constants.StudentStatus(s.Status),
			TeacherDisplay:   strings.Join(teachersByStudent[s.ID], ", "),
			RelatedToDisplay: formatStudentRelationships(relationshipsByStudent[s.ID]),
			CreatedAt:        utils.FormatNullDateTimeSecondsPHT(s.CreatedAt),
			UpdatedAt:        utils.FormatNullDateTimeSecondsPHT(s.UpdatedAt),
		}
	}

	params := listQueryParams(r)
	w.Header().Set("Content-Type", "text/html")
	frontend.Students(frontend.StudentData{
		Students:       viewStudents,
		Query:          q,
		Status:         constants.StudentStatus(status),
		TeacherID:      strconv.FormatInt(teacherID, 10),
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(utils.URL("/students"), page.Number-1, page.Size, params),
		NextURL:        utils.BuildPageURLAt(utils.URL("/students"), page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     utils.URL("/students"),
	}).Render(ctx, w)
}

func handleStudentEdit(w http.ResponseWriter, r *http.Request, studentID int64) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)
	isSuperuser := auth.HasAdminAccess(role)

	existing, err := dbRO.GetQueries().GetStudentByID(ctx, studentID)
	if err != nil {
		HttpError(w, "Student not found", http.StatusNotFound)
		return
	}

	assignedTeachers, err := dbRO.GetQueries().GetTeachersByStudentID(ctx, studentID)
	if err != nil {
		HttpError(w, "Failed to load assigned teachers", http.StatusInternalServerError)
		return
	}

	if !isSuperuser {
		if user.ID == 0 {
			HttpError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := requireStudentAssignedToTeacher(ctx, user.ID, studentID); err != nil {
			HttpError(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	if r.Method == http.MethodGet {
		data, err := studentEditStudentData(ctx, studentID, false)
		if err != nil {
			HttpError(w, "Student not found", http.StatusNotFound)
			return
		}
		data.IsSuperuser = isSuperuser

		w.Header().Set("Content-Type", "text/html")
		frontend.EditStudent(data).Render(ctx, w)
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

	ratePerClass, err := requireFloat64(r.FormValue("ratePerClass"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	teachersBefore := teacherIDsString(assignedTeachers)

	if isSuperuser {
		teacherIDs, err := parseAssignedTeacherIDs(ctx, dbRO.GetQueries(), r.Form["teachers"])
		if err != nil {
			sendErrorLog(w, err.Error())
			return
		}

		req := models.StudentRegisterRequest{
			Name:           r.FormValue("name"),
			Currency:       r.FormValue("currency"),
			Contact:        r.FormValue("contact"),
			RatePerClass:   ratePerClass,
			ParentName:     r.FormValue("parentName"),
			AssignedColor:  r.FormValue("assignedColor"),
			Status:         r.FormValue("status"),
			InactiveReason: r.FormValue("inactiveReason"),
		}
		if err := validateStudentRequest(&req); err != nil {
			sendErrorLog(w, err.Error())
			return
		}

		err = dbRW.GetQueries().UpdateStudent(ctx, queries.UpdateStudentParams{
			Name:           req.Name,
			Currency:       req.Currency,
			Contact:        sql.NullString{String: req.Contact, Valid: req.Contact != ""},
			RatePerClass:   req.RatePerClass,
			ParentName:     sql.NullString{String: req.ParentName, Valid: req.ParentName != ""},
			AssignedColor:  req.AssignedColor,
			Status:         req.Status,
			InactiveReason: sql.NullString{String: req.InactiveReason, Valid: req.InactiveReason != ""},
			ID:             studentID,
		})
		if err != nil {
			sendErrorLog(w, err.Error())
			return
		}

		if err := saveStudentRelationships(ctx, studentID, r); err != nil {
			sendErrorLog(w, err.Error())
			return
		}

		if err := dbRW.GetQueries().DeleteTeacherStudentLinksByStudentID(ctx, studentID); err != nil {
			sendErrorLog(w, err.Error())
			return
		}

		var newTeacherIDs []string
		for _, tid := range teacherIDs {
			if err := dbRW.GetQueries().InsertTeacherStudentM2M(ctx, queries.InsertTeacherStudentM2MParams{
				TeacherID: tid,
				StudentID: studentID,
			}); err != nil {
				sendErrorLog(w, err.Error())
				return
			}
			newTeacherIDs = append(newTeacherIDs, strconv.FormatInt(tid, 10))
		}

		updated, _ := dbRW.GetQueries().GetStudentByID(ctx, studentID)
		insertAuditLogAs(ctx, user, "students", formatStudentAudit(existing, updated, teachersBefore, strings.Join(newTeacherIDs, ",")))
		beforeIDs := make([]int64, 0, len(assignedTeachers))
		for _, t := range assignedTeachers {
			beforeIDs = append(beforeIDs, t.ID)
		}
		notifyNewlyAssignedTeachers(ctx, user, req.Name, beforeIDs, teacherIDs)

		if _, err := fmt.Fprint(w, "Student updated successfully!\n"); err != nil {
			sendErrorLog(w, err.Error())
		}
		return
	}

	inactiveReason := ""
	if existing.InactiveReason.Valid {
		inactiveReason = existing.InactiveReason.String
	}

	req := models.StudentRegisterRequest{
		Name:           r.FormValue("name"),
		Currency:       r.FormValue("currency"),
		RatePerClass:   ratePerClass,
		ParentName:     r.FormValue("parentName"),
		AssignedColor:  r.FormValue("assignedColor"),
		Status:         existing.Status,
		InactiveReason: inactiveReason,
	}
	if err := validateStudentRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	err = dbRW.GetQueries().UpdateStudent(ctx, queries.UpdateStudentParams{
		Name:           req.Name,
		Currency:       req.Currency,
		Contact:        existing.Contact,
		RatePerClass:   req.RatePerClass,
		ParentName:     sql.NullString{String: req.ParentName, Valid: req.ParentName != ""},
		AssignedColor:  req.AssignedColor,
		Status:         existing.Status,
		InactiveReason: existing.InactiveReason,
		ID:             studentID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	updated, _ := dbRW.GetQueries().GetStudentByID(ctx, studentID)
	insertAuditLogAs(ctx, user, "students", formatStudentAudit(existing, updated, teachersBefore, teachersBefore))

	if _, err := fmt.Fprint(w, "Student updated successfully!\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
}
