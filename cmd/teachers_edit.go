package cmd

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/database/queries"
	"zion-english/internal/models"
	"zion-english/internal/utils"
)

func teacherFilterParams(q, status, email string) queries.CountTeachersFilteredParams {
	return queries.CountTeachersFilteredParams{
		Column1: q,
		Column2: sql.NullString{String: q, Valid: true},
		Column3: status,
		Status:  status,
		Column5: email,
		Column6: sql.NullString{String: email, Valid: true},
	}
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

	w.Header().Set("Content-Type", "text/html")
	frontend.TeacherViewModal(frontend.TeacherViewData{
		ID:             strconv.FormatInt(teacherID, 10),
		Name:           row.Name,
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
		Status:         row.Status,
		Avatar:         buildTeacherAvatarProps(row),
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
	email := r.URL.Query().Get("email")
	page := utils.ParsePageQuery(r)

	filter := teacherFilterParams(q, status, email)
	total, err := dbRO.GetQueries().CountTeachersFiltered(ctx, filter)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count teachers: %v", err), http.StatusInternalServerError)
		return
	}
	page.Total = total

	teachers, err := dbRO.GetQueries().GetTeachersFiltered(ctx, queries.GetTeachersFilteredParams{
		Column1: filter.Column1,
		Column2: filter.Column2,
		Column3: filter.Column3,
		Status:  filter.Status,
		Column5: filter.Column5,
		Column6: filter.Column6,
		Limit:   int64(page.Size),
		Offset:  int64(page.Offset()),
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch teachers: %v", err), http.StatusInternalServerError)
		return
	}

	viewTeachers := make([]frontend.TeacherItem, len(teachers))
	for i, t := range teachers {
		viewTeachers[i] = frontend.TeacherItem{
			ID:             strconv.FormatInt(t.ID, 10),
			Name:           t.Name,
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
			Status:         t.Status,
			CreatedAt:      t.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		}
	}

	params := listQueryParams(r)
	w.Header().Set("Content-Type", "text/html")
	frontend.Teachers(frontend.TeacherData{
		Teachers:       viewTeachers,
		Query:          q,
		Status:         status,
		Email:          email,
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        pageURL(utils.URL("/teachers"), page.Number-1, page.Size, params),
		NextURL:        pageURL(utils.URL("/teachers"), page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     utils.URL("/teachers"),
	}).Render(ctx, w)
}

func handleTeacherEdit(w http.ResponseWriter, r *http.Request, teacherID int64) {
	ctx := r.Context()

	existing, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		HttpError(w, "Teacher not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
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

	req := models.TeacherRegisterRequest{
		FirstName:      strings.TrimSpace(r.FormValue("firstName")),
		MiddleName:     strings.TrimSpace(r.FormValue("middleName")),
		LastName:       strings.TrimSpace(r.FormValue("lastName")),
		Birthdate:      r.FormValue("birthdate"),
		Address:        strings.TrimSpace(r.FormValue("address")),
		JoiningDate:    r.FormValue("joiningDate"),
		MobileNumber:   strings.TrimSpace(r.FormValue("mobileNumber")),
		Email:          normalizeEmail(r.FormValue("email")),
		Certifications: r.FormValue("certifications"),
		AssignedColor:  r.FormValue("assignedColor"),
		RatePerClass:   ratePerClass,
		Currency:       r.FormValue("currency"),
		DriveUrl:       r.FormValue("driveUrl"),
		Sex:            r.FormValue("sex"),
	}
	req.Name = utils.ComposePersonName(req.FirstName, req.MiddleName, req.LastName)

	if err := validateTeacherRequest(&req); err != nil {
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

	err = dbRW.GetQueries().UpdateTeacherBySuperuser(ctx, queries.UpdateTeacherBySuperuserParams{
		Name:           req.Name,
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
		ID:             teacherID,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLog(ctx, "teachers", fmt.Sprintf("updated teacher '%s' (id %d)", req.Name, teacherID))

	if _, err := fmt.Fprint(w, "Teacher updated successfully!\n"); err != nil {
		sendErrorLog(w, err.Error())
	}
}
