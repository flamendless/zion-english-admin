package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/trainingmaterials"
	"zion-english/internal/utils"
)

type trainingMaterialRow struct {
	ID              int64
	Title           string
	Description     string
	Url             string
	EmbedUrl        string
	SourceType      string
	VideoID         string
	ThumbnailUrl    string
	DurationSeconds sql.NullInt64
	Status          string
	Required        bool
	CreatedBy       int64
	CreatedAt       string
	UpdatedAt       string
}

func handleTrainingMaterials(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	filters := parseMaterialLibraryFilters(r)
	sort := parseListSort(r, frontend.ListSortKindTrainingMaterial)
	page := utils.ParsePageQuery(r)

	var rows []trainingMaterialRow
	var err error
	if auth.HasAdminAccess(user.Role) {
		total, err := dbRO.GetQueries().CountTrainingMaterialsForAdmin(ctx)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count materials: %v", err), http.StatusInternalServerError)
			return
		}
		adminRows, err := dbRO.GetQueries().GetTrainingMaterialsPagedForAdmin(ctx, queries.GetTrainingMaterialsPagedForAdminParams{
			Limit:  total,
			Offset: 0,
		})
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load materials: %v", err), http.StatusInternalServerError)
			return
		}
		rows = mapTrainingMaterialAdminRows(adminRows)
	} else {
		total, err := dbRO.GetQueries().CountTrainingMaterialsPublished(ctx)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count materials: %v", err), http.StatusInternalServerError)
			return
		}
		publishedRows, err := dbRO.GetQueries().GetTrainingMaterialsPagedPublished(ctx, queries.GetTrainingMaterialsPagedPublishedParams{
			Limit:  total,
			Offset: 0,
		})
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load materials: %v", err), http.StatusInternalServerError)
			return
		}
		rows = mapTrainingMaterialPublishedRows(publishedRows)
	}

	progressByMaterial := map[int64]frontend.TrainingMaterialProgressSummary{}
	if user.Role == auth.RoleTeacher && len(rows) > 0 {
		materialIDs := make([]int64, 0, len(rows))
		for _, row := range rows {
			materialIDs = append(materialIDs, row.ID)
		}
		progressRows, err := dbRO.GetQueries().GetTrainingMaterialProgressByTeacherID(ctx, queries.GetTrainingMaterialProgressByTeacherIDParams{
			TeacherID:   user.ID,
			MaterialIds: materialIDs,
		})
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load progress: %v", err), http.StatusInternalServerError)
			return
		}
		for _, p := range progressRows {
			progressByMaterial[p.MaterialID] = mapTrainingMaterialProgressSummary(p)
		}
	}

	materialIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		materialIDs = append(materialIDs, row.ID)
	}
	tagsByMaterial, err := loadMaterialTagIDsByMaterial(ctx, materialIDs, true)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load tags: %v", err), http.StatusInternalServerError)
		return
	}
	rows = filterTrainingMaterialRows(rows, tagsByMaterial, progressByMaterial, filters, user.Role == auth.RoleTeacher)
	sortTrainingMaterialRows(rows, sort)
	page.Total = int64(len(rows))
	rows = paginateSlice(rows, page)

	items, err := buildTrainingMaterialListItems(ctx, rows, user, progressByMaterial)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	existingTags, err := dbRO.GetQueries().GetAllTrainingMaterialTags(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load tags: %v", err), http.StatusInternalServerError)
		return
	}

	filterPath := utils.URL("/training-materials")
	filterParams := materialLibraryFilterParams(filters, sort)
	data := frontend.TrainingMaterialsData{
		Materials:        items,
		ExistingTags:     mapTrainingMaterialTags(existingTags),
		CanCreate:        trainingmaterials.CanCreate(user.Role),
		CanViewReport:    auth.HasAdminAccess(user.Role),
		Query:            filters.Query,
		StatusFilter:     filters.Status,
		TagFilter:        materialTagFilterValue(filters.TagID),
		ProgressFilter:   filters.Progress,
		SortBy:           sort.By,
		SortOrder:        string(sort.Order),
		ShowStatusFilter: auth.HasAdminAccess(user.Role),
		ShowProgressFilter: user.Role == auth.RoleTeacher,
		FilterPath:       filterPath,
		PageNumber:       page.Number,
		PageTotalPages:   page.TotalPages(),
		PageTotal:        page.Total,
		PrevURL:          utils.BuildPageURLAt(filterPath, page.Number-1, page.Size, filterParams),
		NextURL:          utils.BuildPageURLAt(filterPath, page.Number+1, page.Size, filterParams),
		HasPrev:          page.HasPrev(),
		HasNext:          page.HasNext(),
	}

	if err := frontend.TrainingMaterials(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTrainingMaterialsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "training-materials", "/watch"); ok {
		handleTrainingMaterialWatch(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "training-materials", "/progress"); ok {
		handleTrainingMaterialProgress(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "training-materials", "/view"); ok {
		handleTrainingMaterialView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "training-materials", "/edit"); ok {
		handleTrainingMaterialEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "training-materials", "/delete"); ok {
		handleTrainingMaterialDelete(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleTrainingMaterialCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !trainingmaterials.CanCreate(user.Role) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/training-materials")
		return
	}

	req := parseTrainingMaterialRequest(r)
	if err := trainingmaterials.ValidateRequest(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/training-materials")
		return
	}

	parsed, err := trainingmaterials.ParseURL(trainingmaterials.SourceType(req.SourceType), req.URL)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/training-materials")
		return
	}

	id, err := dbRW.GetQueries().InsertTrainingMaterial(ctx, queries.InsertTrainingMaterialParams{
		Title:        strings.TrimSpace(req.Title),
		Description:  strings.TrimSpace(req.Description),
		Url:          strings.TrimSpace(req.URL),
		EmbedUrl:     parsed.EmbedURL,
		SourceType:   string(parsed.SourceType),
		VideoID:      parsed.VideoID,
		ThumbnailUrl: parsed.ThumbnailURL,
		Status:       req.Status,
		Required:     trialClassToInt64(req.Required),
		CreatedBy:    user.ID,
	})
	if err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to create material: %v", err))
		HttpRedirect(w, r, "/training-materials")
		return
	}

	if err := trainingmaterials.ReplaceMaterialTags(ctx, dbRW.GetQueries(), id, req.TagLabels); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/training-materials")
		return
	}

	insertAuditLogAs(ctx, user, "training-materials", fmt.Sprintf("Created training material #%d", id))
	setSuccessFlash(w, "Training material created successfully")
	HttpRedirect(w, r, "/training-materials")
}

func handleTrainingMaterialView(w http.ResponseWriter, r *http.Request, materialID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	material, tags, err := loadTrainingMaterialWithTags(ctx, materialID)
	if err != nil {
		HttpError(w, "Training material not found", http.StatusNotFound)
		return
	}
	if !trainingmaterials.CanView(user.Role, material.Status) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}

	data := frontend.TrainingMaterialViewData{
		ID:           strconv.FormatInt(materialID, 10),
		Title:        material.Title,
		Description:  material.Description,
		URL:          material.Url,
		ThumbnailURL: material.ThumbnailUrl,
		SourceType:   material.SourceType,
		Status:       material.Status,
		Required:     material.Required == 1,
		CreatedAt:    material.CreatedAt,
		UpdatedAt:    material.UpdatedAt,
		DeletedAt:    material.DeletedAt.String,
		Tags:         mapTrainingMaterialTags(tags),
		CanEdit:      trainingmaterials.CanEdit(user.Role),
		CanWatch:     trainingmaterials.CanWatch(user.Role, material.Status),
	}

	if err := frontend.TrainingMaterialViewModal(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTrainingMaterialEdit(w http.ResponseWriter, r *http.Request, materialID int64) {
	ctx := r.Context()
	user := auth.GetUser(ctx)

	if !trainingmaterials.CanEdit(user.Role) {
		if r.Method == http.MethodGet {
			HttpError(w, MsgForbidden, http.StatusForbidden)
		} else {
			setErrorFlash(w, "You do not have permission to edit training materials")
			HttpRedirect(w, r, "/training-materials")
		}
		return
	}

	material, err := dbRO.GetQueries().GetTrainingMaterialByID(ctx, materialID)
	if err != nil {
		if r.Method == http.MethodGet {
			HttpError(w, "Training material not found", http.StatusNotFound)
		} else {
			setErrorFlash(w, "Training material not found")
			HttpRedirect(w, r, "/training-materials")
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		tags, err := dbRO.GetQueries().GetTagsByTrainingMaterialID(ctx, materialID)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load tags: %v", err), http.StatusInternalServerError)
			return
		}
		existingTags, err := dbRO.GetQueries().GetAllTrainingMaterialTags(ctx)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load tags: %v", err), http.StatusInternalServerError)
			return
		}

		data := frontend.TrainingMaterialFormData{
			ID:           strconv.FormatInt(materialID, 10),
			Title:        material.Title,
			Description:  material.Description,
			URL:          material.Url,
			ThumbnailURL: material.ThumbnailUrl,
			SourceType:   material.SourceType,
			Status:       material.Status,
			Required:     material.Required == 1,
			SelectedTags: mapTrainingMaterialTags(tags),
			ExistingTags: mapTrainingMaterialTags(existingTags),
			IsEdit:       true,
			IsDeleted:    material.Status == trainingmaterials.StatusDeleted,
			CanDelete:    trainingmaterials.CanDelete(user.Role),
		}
		if err := frontend.TrainingMaterialFormModal(data).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			setErrorFlash(w, "Invalid form data")
			HttpRedirect(w, r, "/training-materials")
			return
		}
		req := parseTrainingMaterialRequest(r)
		if err := trainingmaterials.ValidateEditRequest(req); err != nil {
			setErrorFlash(w, err.Error())
			HttpRedirect(w, r, "/training-materials")
			return
		}
		if req.Status == trainingmaterials.StatusDeleted && !trainingmaterials.CanDelete(user.Role) {
			setErrorFlash(w, "You do not have permission to delete this material")
			HttpRedirect(w, r, "/training-materials")
			return
		}

		parsed, err := trainingmaterials.ParseURL(trainingmaterials.SourceType(req.SourceType), req.URL)
		if err != nil {
			setErrorFlash(w, err.Error())
			HttpRedirect(w, r, "/training-materials")
			return
		}

		if err := dbRW.GetQueries().UpdateTrainingMaterial(ctx, queries.UpdateTrainingMaterialParams{
			Title:        strings.TrimSpace(req.Title),
			Description:  strings.TrimSpace(req.Description),
			Url:          strings.TrimSpace(req.URL),
			EmbedUrl:     parsed.EmbedURL,
			SourceType:   string(parsed.SourceType),
			VideoID:      parsed.VideoID,
			ThumbnailUrl: parsed.ThumbnailURL,
			Status:       req.Status,
			Required:     trialClassToInt64(req.Required),
			ID:           materialID,
		}); err != nil {
			setErrorFlash(w, fmt.Sprintf("Failed to update material: %v", err))
			HttpRedirect(w, r, "/training-materials")
			return
		}
		if err := trainingmaterials.ReplaceMaterialTags(ctx, dbRW.GetQueries(), materialID, req.TagLabels); err != nil {
			setErrorFlash(w, err.Error())
			HttpRedirect(w, r, "/training-materials")
			return
		}
		insertAuditLogAs(ctx, user, "training-materials", fmt.Sprintf("Updated training material #%d", materialID))
		setSuccessFlash(w, "Training material updated successfully")
		HttpRedirect(w, r, "/training-materials")
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleTrainingMaterialDelete(w http.ResponseWriter, r *http.Request, materialID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !trainingmaterials.CanDelete(user.Role) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}

	material, err := dbRO.GetQueries().GetTrainingMaterialByID(ctx, materialID)
	if err != nil {
		HttpError(w, "Training material not found", http.StatusNotFound)
		return
	}
	if material.Status == trainingmaterials.StatusDeleted {
		setErrorFlash(w, "Training material is already deleted")
		HttpRedirect(w, r, "/training-materials")
		return
	}

	if err := dbRW.GetQueries().DeleteTrainingMaterial(ctx, materialID); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to delete material: %v", err))
		HttpRedirect(w, r, "/training-materials")
		return
	}

	insertAuditLogAs(ctx, user, "training-materials", fmt.Sprintf("Deleted training material #%d", materialID))
	setSuccessFlash(w, "Training material deleted successfully")
	HttpRedirect(w, r, "/training-materials")
}

func handleTrainingMaterialWatch(w http.ResponseWriter, r *http.Request, materialID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	material, tags, err := loadTrainingMaterialWithTags(ctx, materialID)
	if err != nil {
		HttpError(w, "Training material not found", http.StatusNotFound)
		return
	}
	if !trainingmaterials.CanWatch(user.Role, material.Status) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}

	var progress frontend.TrainingMaterialProgressSummary
	if user.Role == auth.RoleTeacher {
		p, err := dbRO.GetQueries().GetTrainingMaterialProgressByTeacherAndMaterial(ctx, queries.GetTrainingMaterialProgressByTeacherAndMaterialParams{
			MaterialID: materialID,
			TeacherID:  user.ID,
		})
		if err == nil {
			progress = mapTrainingMaterialProgressSummary(p)
		}
	}

	durationSeconds := int64(0)
	if material.DurationSeconds.Valid {
		durationSeconds = material.DurationSeconds.Int64
	}

	data := frontend.TrainingMaterialWatchData{
		ID:              strconv.FormatInt(materialID, 10),
		Title:           material.Title,
		Description:     material.Description,
		EmbedURL:        material.EmbedUrl,
		VideoID:         material.VideoID,
		SourceType:      material.SourceType,
		ThumbnailURL:    material.ThumbnailUrl,
		Tags:            mapTrainingMaterialTags(tags),
		DurationSeconds: durationSeconds,
		Progress:        progress,
		ProgressURL:     utils.URL("/training-materials/" + strconv.FormatInt(materialID, 10) + "/progress"),
		BackURL:         utils.URL("/training-materials"),
	}

	if err := frontend.TrainingMaterialWatch(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTrainingMaterialProgress(w http.ResponseWriter, r *http.Request, materialID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user.Role != auth.RoleTeacher {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}

	material, err := dbRO.GetQueries().GetTrainingMaterialByID(ctx, materialID)
	if err != nil {
		HttpError(w, "Training material not found", http.StatusNotFound)
		return
	}
	if !trainingmaterials.CanWatch(user.Role, material.Status) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}

	var payload struct {
		WatchSeconds    int64   `json:"watch_seconds"`
		ProgressPercent float64 `json:"progress_percent"`
		DurationSeconds int64   `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		HttpError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.DurationSeconds > 0 && (!material.DurationSeconds.Valid || material.DurationSeconds.Int64 == 0) {
		if err := dbRW.GetQueries().UpdateTrainingMaterialDuration(ctx, queries.UpdateTrainingMaterialDurationParams{
			DurationSeconds: sql.NullInt64{Int64: payload.DurationSeconds, Valid: true},
			ID:              materialID,
		}); err != nil {
			HttpError(w, fmt.Sprintf("Failed to update duration: %v", err), http.StatusInternalServerError)
			return
		}
		material.DurationSeconds = sql.NullInt64{Int64: payload.DurationSeconds, Valid: true}
	}

	durationSeconds := payload.DurationSeconds
	if durationSeconds <= 0 && material.DurationSeconds.Valid {
		durationSeconds = material.DurationSeconds.Int64
	}

	existing, hasExisting := loadTrainingMaterialProgress(ctx, materialID, user.ID)
	watchSeconds := payload.WatchSeconds
	if hasExisting && watchSeconds < existing.WatchSeconds {
		watchSeconds = existing.WatchSeconds
	}
	if durationSeconds > 0 && watchSeconds > durationSeconds {
		watchSeconds = durationSeconds
	}

	progressPercent := payload.ProgressPercent
	if durationSeconds > 0 {
		progressPercent = trainingmaterials.ComputeProgress(watchSeconds, durationSeconds)
	}

	var completedAt sql.NullString
	if trainingmaterials.IsCompleted(progressPercent) {
		completedAt = sql.NullString{String: time.Now().UTC().Format(constants.DateTimeSecondsLayout), Valid: true}
	}
	if hasExisting && existing.CompletedAt.Valid {
		completedAt = existing.CompletedAt
	}

	if err := dbRW.GetQueries().UpsertTrainingMaterialProgress(ctx, queries.UpsertTrainingMaterialProgressParams{
		MaterialID:      materialID,
		TeacherID:       user.ID,
		WatchSeconds:    watchSeconds,
		ProgressPercent: progressPercent,
		CompletedAt:     completedAt,
	}); err != nil {
		HttpError(w, fmt.Sprintf("Failed to save progress: %v", err), http.StatusInternalServerError)
		return
	}

	if !hasExisting {
		insertAuditLogAs(ctx, user, "training-materials", fmt.Sprintf("Started training material #%d", materialID))
	}

	writeJSON(w)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"progress_percent": progressPercent,
		"completed":        completedAt.Valid,
	})
}

func handleTrainingMaterialURLPreview(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	sourceType, err := trainingmaterials.InferSourceType(rawURL)
	if err != nil {
		writeJSON(w)
		_ = json.NewEncoder(w).Encode(map[string]string{"thumbnail_url": "", "error": err.Error()})
		return
	}
	parsed, err := trainingmaterials.ParseURL(sourceType, rawURL)
	writeJSON(w)
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]string{"thumbnail_url": ""})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{
		"thumbnail_url": parsed.ThumbnailURL,
		"source_type":   string(parsed.SourceType),
		"video_id":      parsed.VideoID,
	})
}

func handleTrainingMaterialsProgressReport(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(user.Role) {
		HttpError(w, MsgForbidden, http.StatusForbidden)
		return
	}

	page := utils.ParsePageQuery(r)
	materialIDStr := strings.TrimSpace(r.URL.Query().Get("material_id"))
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	if statusFilter == "" {
		statusFilter = "all"
	}

	var materialID int64
	if materialIDStr != "" {
		materialID, _ = strconv.ParseInt(materialIDStr, 10, 64)
	}

	countParams := queries.CountTrainingMaterialProgressReportParams{
		Column1:    materialID,
		MaterialID: materialID,
		Column3:    statusFilter,
		Column4:    statusFilter,
		Column5:    statusFilter,
		Column6:    statusFilter,
	}
	var err error
	page.Total, err = dbRO.GetQueries().CountTrainingMaterialProgressReport(ctx, countParams)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count progress: %v", err), http.StatusInternalServerError)
		return
	}

	reportParams := queries.GetTrainingMaterialProgressReportParams{
		Column1:    materialID,
		MaterialID: materialID,
		Column3:    statusFilter,
		Column4:    statusFilter,
		Column5:    statusFilter,
		Column6:    statusFilter,
		Limit:      int64(page.Size),
		Offset:     int64(page.Offset()),
	}
	rows, err := dbRO.GetQueries().GetTrainingMaterialProgressReport(ctx, reportParams)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load progress: %v", err), http.StatusInternalServerError)
		return
	}

	materials, err := dbRO.GetQueries().GetAllTrainingMaterialsForSelect(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load materials: %v", err), http.StatusInternalServerError)
		return
	}

	filterPath := utils.URL("/training-materials/progress")
	filterParams := map[string]string{
		"material_id": materialIDStr,
		"status":      statusFilter,
	}

	items := make([]frontend.TrainingMaterialProgressReportItem, 0, len(rows))
	for _, row := range rows {
		teacherName := utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName)
		avatar := buildTeacherListAvatarProps(row.TeacherID, row.FirstName, row.MiddleName, row.LastName, row.AssignedColor, row.ProfilePicture)
		items = append(items, frontend.TrainingMaterialProgressReportItem{
			MaterialID:      strconv.FormatInt(row.MaterialID, 10),
			MaterialTitle:   row.MaterialTitle,
			TeacherID:       strconv.FormatInt(row.TeacherID, 10),
			TeacherName:     teacherName,
			TeacherAvatar:   avatar,
			ProgressPercent: row.ProgressPercent,
			CompletedAt:     row.CompletedAt.String,
			LastViewedAt:    row.LastViewedAt,
			IsCompleted:     row.CompletedAt.Valid,
		})
	}

	materialOptions := make([]frontend.TrainingMaterialSelectOption, 0, len(materials))
	for _, m := range materials {
		materialOptions = append(materialOptions, frontend.TrainingMaterialSelectOption{
			ID:    strconv.FormatInt(m.ID, 10),
			Title: m.Title,
		})
	}

	data := frontend.TrainingMaterialsProgressData{
		Items:           items,
		MaterialOptions: materialOptions,
		SelectedMaterial: materialIDStr,
		StatusFilter:    statusFilter,
		BackURL:         utils.URL("/training-materials"),
		PageNumber:      page.Number,
		PageTotalPages:  page.TotalPages(),
		PageTotal:       page.Total,
		PrevURL:         utils.BuildPageURLAt(filterPath, page.Number-1, page.Size, filterParams),
		NextURL:         utils.BuildPageURLAt(filterPath, page.Number+1, page.Size, filterParams),
		HasPrev:         page.HasPrev(),
		HasNext:         page.HasNext(),
	}

	if err := frontend.TrainingMaterialsProgressReport(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseTrainingMaterialRequest(r *http.Request) trainingmaterials.Request {
	status := r.FormValue("status")
	if status == "" {
		status = trainingmaterials.StatusDraft
	}
	sourceType := r.FormValue("source_type")
	if sourceType == "" {
		sourceType = string(trainingmaterials.SourceYouTube)
	}
	return trainingmaterials.Request{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		URL:         r.FormValue("url"),
		SourceType:  sourceType,
		Status:      status,
		Required:    r.FormValue("required") == "1",
		TagLabels:   r.Form["tags"],
	}
}

func loadTrainingMaterialWithTags(ctx context.Context, materialID int64) (queries.GetTrainingMaterialByIDRow, []queries.TblTrainingMaterialTag, error) {
	material, err := dbRO.GetQueries().GetTrainingMaterialByID(ctx, materialID)
	if err != nil {
		return queries.GetTrainingMaterialByIDRow{}, nil, err
	}
	tags, err := dbRO.GetQueries().GetTagsByTrainingMaterialID(ctx, materialID)
	if err != nil {
		return queries.GetTrainingMaterialByIDRow{}, nil, err
	}
	return material, tags, nil
}

func loadTrainingMaterialProgress(ctx context.Context, materialID, teacherID int64) (queries.TblTrainingMaterialProgress, bool) {
	row, err := dbRO.GetQueries().GetTrainingMaterialProgressByTeacherAndMaterial(ctx, queries.GetTrainingMaterialProgressByTeacherAndMaterialParams{
		MaterialID: materialID,
		TeacherID:  teacherID,
	})
	if err != nil {
		return queries.TblTrainingMaterialProgress{}, false
	}
	return row, true
}

func mapTrainingMaterialAdminRows(rows []queries.GetTrainingMaterialsPagedForAdminRow) []trainingMaterialRow {
	out := make([]trainingMaterialRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, trainingMaterialRowFromQuery(trainingMaterialQueryRow{
			ID:              row.ID,
			Title:           row.Title,
			Description:     row.Description,
			Url:             row.Url,
			EmbedUrl:        row.EmbedUrl,
			SourceType:      row.SourceType,
			VideoID:         row.VideoID,
			ThumbnailUrl:    row.ThumbnailUrl,
			DurationSeconds: row.DurationSeconds,
			Status:          row.Status,
			Required:        row.Required,
			CreatedBy:       row.CreatedBy,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		}))
	}
	return out
}

func mapTrainingMaterialPublishedRows(rows []queries.GetTrainingMaterialsPagedPublishedRow) []trainingMaterialRow {
	out := make([]trainingMaterialRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, trainingMaterialRowFromQuery(trainingMaterialQueryRow{
			ID:              row.ID,
			Title:           row.Title,
			Description:     row.Description,
			Url:             row.Url,
			EmbedUrl:        row.EmbedUrl,
			SourceType:      row.SourceType,
			VideoID:         row.VideoID,
			ThumbnailUrl:    row.ThumbnailUrl,
			DurationSeconds: row.DurationSeconds,
			Status:          row.Status,
			Required:        row.Required,
			CreatedBy:       row.CreatedBy,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		}))
	}
	return out
}

type trainingMaterialQueryRow struct {
	ID              int64
	Title           string
	Description     string
	Url             string
	EmbedUrl        string
	SourceType      string
	VideoID         string
	ThumbnailUrl    string
	DurationSeconds sql.NullInt64
	Status          string
	Required        int64
	CreatedBy       int64
	CreatedAt       string
	UpdatedAt       string
}

func trainingMaterialRowFromQuery(row trainingMaterialQueryRow) trainingMaterialRow {
	return trainingMaterialRow{
		ID:              row.ID,
		Title:           row.Title,
		Description:     row.Description,
		Url:             row.Url,
		EmbedUrl:        row.EmbedUrl,
		SourceType:      row.SourceType,
		VideoID:         row.VideoID,
		ThumbnailUrl:    row.ThumbnailUrl,
		DurationSeconds: row.DurationSeconds,
		Status:          row.Status,
		Required:        row.Required == 1,
		CreatedBy:       row.CreatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func buildTrainingMaterialListItems(ctx context.Context, rows []trainingMaterialRow, user auth.User, progressByMaterial map[int64]frontend.TrainingMaterialProgressSummary) ([]frontend.TrainingMaterialListItem, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	materialIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		materialIDs = append(materialIDs, row.ID)
	}

	tagRows, err := dbRO.GetQueries().GetTagsByTrainingMaterialIDs(ctx, materialIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	tagsByMaterial := make(map[int64][]frontend.TrainingMaterialTag)
	for _, tr := range tagRows {
		tagsByMaterial[tr.MaterialID] = append(tagsByMaterial[tr.MaterialID], frontend.TrainingMaterialTag{
			ID:    strconv.FormatInt(tr.ID, 10),
			Label: tr.Label,
			Color: tr.Color,
		})
	}

	items := make([]frontend.TrainingMaterialListItem, 0, len(rows))
	for _, row := range rows {
		progress := progressByMaterial[row.ID]
		items = append(items, frontend.TrainingMaterialListItem{
			ID:                strconv.FormatInt(row.ID, 10),
			Title:             row.Title,
			Description:       row.Description,
			URL:               row.Url,
			ThumbnailURL:      row.ThumbnailUrl,
			SourceType:        row.SourceType,
			Status:            row.Status,
			Required:          row.Required,
			Tags:              tagsByMaterial[row.ID],
			PublicationStatus: trainingMaterialPublicationStatus(row.Status),
			PublicationTone:   trainingMaterialPublicationTone(row.Status),
			IsDeleted:         row.Status == trainingmaterials.StatusDeleted,
			CanEdit:           trainingmaterials.CanEdit(user.Role),
			CanWatch:          trainingmaterials.CanWatch(user.Role, row.Status),
			Progress:          progress,
			HasProgress:       progress.ProgressPercent > 0 || progress.IsCompleted,
		})
	}
	return items, nil
}

func mapTrainingMaterialTags(tags []queries.TblTrainingMaterialTag) []frontend.TrainingMaterialTag {
	out := make([]frontend.TrainingMaterialTag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, frontend.TrainingMaterialTag{
			ID:    strconv.FormatInt(tag.ID, 10),
			Label: tag.Label,
			Color: tag.Color,
		})
	}
	return out
}

func mapTrainingMaterialProgressSummary(p queries.TblTrainingMaterialProgress) frontend.TrainingMaterialProgressSummary {
	return frontend.TrainingMaterialProgressSummary{
		ProgressPercent: p.ProgressPercent,
		IsCompleted:     p.CompletedAt.Valid,
		CompletedLabel:  trainingMaterialProgressLabel(p.ProgressPercent, p.CompletedAt.Valid),
		ProgressTone:    trainingMaterialProgressTone(p.ProgressPercent, p.CompletedAt.Valid),
	}
}

func trainingMaterialPublicationStatus(status string) string {
	switch status {
	case trainingmaterials.StatusPublished:
		return "Published"
	case trainingmaterials.StatusDeleted:
		return "Deleted"
	default:
		return "Draft"
	}
}

func trainingMaterialPublicationTone(status string) frontend.PillTone {
	switch status {
	case trainingmaterials.StatusPublished:
		return frontend.PillToneSuccess
	case trainingmaterials.StatusDeleted:
		return frontend.PillToneError
	default:
		return frontend.PillToneNeutral
	}
}

func trainingMaterialProgressLabel(percent float64, completed bool) string {
	if completed {
		return "Completed"
	}
	if percent <= 0 {
		return "Not started"
	}
	return fmt.Sprintf("%.0f%%", percent)
}

func trainingMaterialProgressTone(percent float64, completed bool) frontend.PillTone {
	if completed {
		return frontend.PillToneSuccess
	}
	if percent <= 0 {
		return frontend.PillToneNeutral
	}
	return frontend.PillToneInfo
}
