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
	HttpError(w, "Not found", http.StatusNotFound)
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
			Status:           s.Status,
			RelatedToDisplay: formatStudentRelationships(relationshipsByStudent[s.ID]),
			CreatedAt:        s.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			UpdatedAt:        s.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
		}
	}

	params := listQueryParams(r)
	w.Header().Set("Content-Type", "text/html")
	frontend.Students(frontend.StudentData{
		Students:       viewStudents,
		Query:          q,
		Status:         status,
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

	if r.Method == http.MethodGet {
		teacherIDs := make([]string, len(assignedTeachers))
		for i, t := range assignedTeachers {
			teacherIDs[i] = strconv.FormatInt(t.ID, 10)
		}

		w.Header().Set("Content-Type", "text/html")
		frontend.EditStudent(frontend.EditStudentData{
			ID:            strconv.FormatInt(studentID, 10),
			Name:          existing.Name,
			Currency:      existing.Currency,
			Contact:       existing.Contact.String,
			RatePerClass:  existing.RatePerClass,
			ParentName:    existing.ParentName.String,
			AssignedColor: existing.AssignedColor,
			Status:        existing.Status,
			TeacherIDs:    teacherIDs,
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

	ratePerClass, err := requireFloat64(r.FormValue("ratePerClass"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	teacherIDs, err := parseAssignedTeacherIDs(ctx, dbRO.GetQueries(), r.Form["teachers"])
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	req := models.StudentRegisterRequest{
		Name:          r.FormValue("name"),
		Currency:      r.FormValue("currency"),
		Contact:       r.FormValue("contact"),
		RatePerClass:  ratePerClass,
		ParentName:    r.FormValue("parentName"),
		AssignedColor: r.FormValue("assignedColor"),
		Status:        r.FormValue("status"),
	}
	if err := validateStudentRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	dup, err := dbRO.GetQueries().GetStudentByNameExcludingID(ctx, queries.GetStudentByNameExcludingIDParams{
		Name: req.Name,
		ID:   studentID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if dup > 0 {
		sendErrorLog(w, "a student with this name already exists")
		return
	}

	teachersBefore := teacherIDsString(assignedTeachers)

	err = dbRW.GetQueries().UpdateStudent(ctx, queries.UpdateStudentParams{
		Name:          req.Name,
		Currency:      req.Currency,
		Contact:       sql.NullString{String: req.Contact, Valid: req.Contact != ""},
		RatePerClass:  req.RatePerClass,
		ParentName:    sql.NullString{String: req.ParentName, Valid: req.ParentName != ""},
		AssignedColor: req.AssignedColor,
		Status:        req.Status,
		ID:            studentID,
	})
	if err != nil {
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
	insertAuditLogAs(ctx, auth.GetUser(ctx), "students", formatStudentAudit(existing, updated, teachersBefore, strings.Join(newTeacherIDs, ",")))

	if _, err := fmt.Fprint(w, "Student updated successfully!\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
}
