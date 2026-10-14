package cmd

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/entitlements"
	"zion-english/internal/utils"
)

func handleHubSelf(w http.ResponseWriter, r *http.Request) {
	handleHubPage(w, r, 0)
}

func handleHubForTeacher(w http.ResponseWriter, r *http.Request, pathTeacherID int64) {
	handleHubPage(w, r, pathTeacherID)
}

func handleHubPartials(w http.ResponseWriter, r *http.Request) {
	const marker = "/hub/partials/"
	idx := strings.Index(r.URL.Path, marker)
	if idx < 0 {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path[idx+len(marker):], "/"), "/")
	switch path {
	case "students":
		handleHubStudentsPartial(w, r)
	case "earnings":
		handleHubEarningsPartial(w, r)
	case "upload-logs":
		handleHubUploadLogsPartial(w, r)
	case "analytics-date-preset":
		handleHubAnalyticsDatePresetPartial(w, r)
	default:
		HttpError(w, MsgNotFound, http.StatusNotFound)
	}
}

func handleHubPage(w http.ResponseWriter, r *http.Request, pathTeacherID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	teacherID, viewAsAdmin, err := resolveHubTeacherID(ctx, r, pathTeacherID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			HttpError(w, MsgForbidden, http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			HttpError(w, MsgUnauthorized, http.StatusUnauthorized)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			HttpError(w, MsgTeacherNotFound, http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			HttpError(w, MsgTeacherNotFound, http.StatusNotFound)
			return
		}
		HttpError(w, "Failed to load teacher", http.StatusInternalServerError)
		return
	}

	teacherName := utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName)
	cutoffStart, cutoffEnd := utils.ActiveCutoffDates()
	resources, err := loadHubResources(ctx, teacherID)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isPro := false
	plan, planErr := entitlements.TeacherPlan(ctx, dbRO.GetQueries(), teacherID)
	if planErr == nil {
		isPro = plan.IsPro
	}

	roleStrings, _ := loadTeacherRoles(ctx, teacherID)
	data := frontend.HubData{
		ViewAsAdmin:     viewAsAdmin,
		TeacherID:       strconv.FormatInt(teacherID, 10),
		TeacherName:     teacherName,
		Avatar:          avatarWithTeacherRoles(buildTeacherAvatarProps(row), roleStrings, constants.TeacherStatus(row.Status)),
		Status:          constants.TeacherStatus(row.Status),
		CutoffStartDate: cutoffStart,
		CutoffEndDate:   cutoffEnd,
		TodayDate:       utils.TodayPHT(),
		Resources:       resources,
		IsPro:           isPro,
	}

	writeHTML(w)
	if err := frontend.Hub(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func resolveHubTeacherID(ctx context.Context, r *http.Request, pathTeacherID int64) (teacherID int64, viewAsAdmin bool, err error) {
	role := auth.GetRole(ctx)
	if auth.HasAdminAccess(role) {
		if pathTeacherID <= 0 {
			return 0, false, ErrInvalidTeacherID
		}
		_, err := dbRO.GetQueries().GetTeacherByID(ctx, pathTeacherID)
		if err != nil {
			return 0, false, err
		}
		return pathTeacherID, true, nil
	}

	if auth.IsTeacherScoped(role) {
		user := auth.GetUser(ctx)
		if user.ID == 0 {
			return 0, false, ErrUnauthorized
		}
		if pathTeacherID > 0 && pathTeacherID != user.ID {
			return 0, false, ErrForbidden
		}
		queryID := strings.TrimSpace(r.URL.Query().Get("teacherId"))
		if queryID != "" {
			parsedID, parseErr := strconv.ParseInt(queryID, 10, 64)
			if parseErr != nil {
				return 0, false, ErrInvalidTeacherID
			}
			if parsedID != user.ID {
				return 0, false, ErrForbidden
			}
		}
		return user.ID, false, nil
	}

	return 0, false, ErrForbidden
}

func resolveHubTeacherIDFromRequest(ctx context.Context, r *http.Request) (int64, error) {
	role := auth.GetRole(ctx)
	if auth.IsTeacherScoped(role) {
		user := auth.GetUser(ctx)
		if user.ID == 0 {
			return 0, ErrUnauthorized
		}
		queryID := strings.TrimSpace(r.URL.Query().Get("teacherId"))
		if queryID != "" {
			parsedID, parseErr := strconv.ParseInt(queryID, 10, 64)
			if parseErr != nil {
				return 0, ErrInvalidTeacherID
			}
			if parsedID != user.ID {
				return 0, ErrForbidden
			}
		}
		return user.ID, nil
	}
	teacherIDStr := strings.TrimSpace(r.URL.Query().Get("teacherId"))
	if teacherIDStr == "" {
		return 0, ErrInvalidTeacherID
	}
	parsedID, err := strconv.ParseInt(teacherIDStr, 10, 64)
	if err != nil || parsedID <= 0 {
		return 0, ErrInvalidTeacherID
	}
	_, err = dbRO.GetQueries().GetTeacherByID(ctx, parsedID)
	if err != nil {
		return 0, err
	}
	return parsedID, nil
}

func loadHubResources(ctx context.Context, teacherID int64) (frontend.HubResourcesData, error) {
	docRows, err := dbRO.GetQueries().GetTeacherDocumentsByTeacherIDFiltered(ctx, documentTeacherFilterParams(teacherID, documentFilters{}))
	if err != nil {
		return frontend.HubResourcesData{}, err
	}
	docs := make([]frontend.HubDocumentSummary, 0, len(docRows))
	for _, row := range docRows {
		docs = append(docs, frontend.HubDocumentSummary{
			Type:     row.Type,
			Status:   constants.TeacherDocumentStatus(row.Status),
			Filename: row.OriginalFilename,
			ViewURL:  utils.URL("/documents/" + strconv.FormatInt(row.ID, 10) + "/file"),
		})
	}

	introRows, err := dbRO.GetQueries().GetTeacherIntroVideosByTeacherIDFiltered(ctx, introVideoTeacherFilterParams(teacherID, introVideoFilters{}))
	if err != nil {
		return frontend.HubResourcesData{}, err
	}
	intro := frontend.HubIntroSummary{}
	if len(introRows) > 0 {
		item := mapIntroVideoItem(introRows[0])
		intro.Status = item.Status
		intro.ViewURL = item.ViewURL
		intro.UploadedAt = item.UploadedAt
	}

	lmRows, err := dbRO.GetQueries().GetLearningMaterialsPagedForUser(ctx, queries.GetLearningMaterialsPagedForUserParams{
		OwnerID: teacherID,
		Limit:   6,
		Offset:  0,
	})
	if err != nil {
		return frontend.HubResourcesData{}, err
	}
	materials := make([]frontend.HubLearningMaterialSummary, 0, len(lmRows))
	for _, row := range lmRows {
		materials = append(materials, frontend.HubLearningMaterialSummary{
			ID:     strconv.FormatInt(row.ID, 10),
			Title:  row.Title,
			Status: row.Status,
		})
	}

	return frontend.HubResourcesData{
		Documents:         docs,
		IntroVideo:        intro,
		LearningMaterials: materials,
	}, nil
}

func handleHubStudentsPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	teacherID, err := resolveHubTeacherIDFromRequest(ctx, r)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			HttpError(w, MsgForbidden, http.StatusForbidden)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	sort := parseListSort(r, frontend.ListSortKindMyStudent)
	page := utils.ParsePageQuery(r)

	rows, page, err := loadMyStudentsPage(ctx, teacherID, q, status, sort, page)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pagination := frontend.BuildPaginationData(page.Number, page.Size, page.Total)
	partialsURL := utils.URL("/hub/partials/students")
	includeSelector := "#hubStudentsToolbar"
	targetSelector := "#hubStudentsTableBody"
	paginationOOBID := "hubStudentsPagination"
	emptyMsg := "No students assigned to this teacher."
	if q != "" || status != "" {
		emptyMsg = "No students found."
	}

	renderHTMLComponent(ctx, w, "render hub students partial", frontend.MyStudentsListPartial(rows, emptyMsg, pagination, partialsURL, includeSelector, targetSelector, paginationOOBID))
}

func loadMyStudentsPage(ctx context.Context, teacherID int64, q, status string, sort utils.SortParams, page utils.Page) ([]frontend.StudentItem, utils.Page, error) {
	filter := queries.CountStudentsByTeacherIDFilteredParams{
		TeacherID: teacherID,
		Column2:   q,
		Column3:   sql.NullString{String: q, Valid: true},
		Column4:   status,
		Status:    status,
	}

	total, err := dbRO.GetQueries().CountStudentsByTeacherIDFiltered(ctx, filter)
	if err != nil {
		return nil, page, err
	}
	page.Total = total

	allStudents, err := dbRO.GetQueries().GetStudentsByTeacherIDFiltered(ctx, queries.GetStudentsByTeacherIDFilteredParams{
		TeacherID: teacherID,
		Column2:   q,
		Column3:   sql.NullString{String: q, Valid: true},
		Column4:   status,
		Status:    status,
		Limit:     total,
		Offset:    0,
	})
	if err != nil {
		return nil, page, err
	}
	sortMyStudentRows(allStudents, sort)
	students := paginateSlice(allStudents, page)

	viewStudents := make([]frontend.StudentItem, len(students))
	for i, s := range students {
		viewStudents[i] = frontend.StudentItem{
			ID:            strconv.FormatInt(s.ID, 10),
			Name:          s.Name,
			StudentNumber: studentNumberString(s.StudentNumber),
			Currency:      s.Currency,
			Contact:       s.Contact.String,
			RatePerClass:  s.RatePerClass,
			ParentName:    s.ParentName.String,
			AssignedColor: s.AssignedColor,
			Status:        constants.StudentStatus(s.Status),
			Avatar:        frontend.StudentAvatarProps(s.Name, s.AssignedColor, false, "sm"),
		}
	}
	return viewStudents, page, nil
}

func handleHubEarningsPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	teacherID, err := resolveHubTeacherIDFromRequest(ctx, r)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			HttpError(w, MsgForbidden, http.StatusForbidden)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, emptyMsg, err := loadEarningsCutoffRows(ctx, teacherID)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderHTMLComponent(ctx, w, "render hub earnings partial", frontend.EarningsCutoffTableBody(rows, emptyMsg))
}

func handleHubUploadLogsPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	teacherID, err := resolveHubTeacherIDFromRequest(ctx, r)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			HttpError(w, MsgForbidden, http.StatusForbidden)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sort := parseListSort(r, frontend.ListSortKindUploadLog)
	page := utils.ParsePageQuery(r)
	filters := parseLogFilters(r)

	countParams := countUploadLogsByTeacherParams(teacherID, filters)
	total, err := dbRO.GetQueries().CountUploadLogsByCreatedByFiltered(ctx, countParams)
	if err != nil {
		HttpError(w, "Failed to count upload logs", http.StatusInternalServerError)
		return
	}
	page.Total = total

	allRows, err := dbRO.GetQueries().GetUploadLogsByCreatedByFiltered(ctx, getUploadLogsByTeacherParams(teacherID, filters, total, 0))
	if err != nil {
		HttpError(w, "Failed to fetch upload logs", http.StatusInternalServerError)
		return
	}
	sortUploadLogByUserRows(allRows, sort)
	rows := paginateSlice(allRows, page)
	viewLogs := uploadLogItemsFromTeacherRows(ctx, rows)

	pagination := frontend.BuildPaginationData(page.Number, page.Size, page.Total)
	partialsURL := utils.URL("/hub/partials/upload-logs")
	includeSelector := "#hubUploadLogsToolbar"
	targetSelector := "#hubUploadLogsTableBody"
	paginationOOBID := "hubUploadLogsPagination"
	emptyMsg := "No upload logs recorded yet."
	if filters.module != "" || filters.message != "" || filters.startDate != "" || filters.endDate != "" {
		emptyMsg = "No upload logs match your filters."
	}

	renderHTMLComponent(ctx, w, "render hub upload logs partial", frontend.HubUploadLogsPartial(viewLogs, emptyMsg, pagination, partialsURL, includeSelector, targetSelector, paginationOOBID))
}

func handleHubAnalyticsDatePresetPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	groupID := strings.TrimSpace(r.URL.Query().Get("groupId"))
	selectID := strings.TrimSpace(r.URL.Query().Get("selectId"))
	if groupID == "" {
		groupID = "hubDatePresetGroup"
	}
	if selectID == "" {
		selectID = "hubDatePreset"
	}
	writeHTML(w)
	if err := frontend.DatePresetForMonthScoped(month, true, groupID, selectID).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}
