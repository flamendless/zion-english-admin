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
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/notifications"
	"zion-english/internal/storage"
	"zion-english/internal/teacherintrovideo"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func introVideoLinkLabel(url sql.NullString, filename sql.NullString) string {
	if url.Valid && url.String != "" {
		return url.String
	}
	if filename.Valid && filename.String != "" {
		return filename.String
	}
	return "-"
}

func introVideoViewURL(id int64, url sql.NullString) string {
	if url.Valid && url.String != "" {
		return url.String
	}
	return utils.URL(fmt.Sprintf("/intro-videos/%d/file", id))
}

func introVideoSourceType(sourceType sql.NullString) constants.TeacherIntroVideoSourceType {
	if !sourceType.Valid {
		return ""
	}
	return constants.TeacherIntroVideoSourceType(sourceType.String)
}

func mapIntroVideoItem(row queries.TblTeacherIntroVideo) frontend.IntroVideoItem {
	linkLabel := introVideoLinkLabel(row.Url, row.OriginalFilename)
	return frontend.IntroVideoItem{
		ID:           strconv.FormatInt(row.ID, 10),
		LinkLabel:    linkLabel,
		URL:          nullStringValue(row.Url),
		SourceType:   introVideoSourceType(row.SourceType),
		Status:       constants.TeacherIntroVideoStatus(row.Status),
		UploadedAt:   utils.FormatNullDateTimePHT(row.CreatedAt),
		RejectReason: nullStringValue(row.RejectReason),
		ViewURL:      introVideoViewURL(row.ID, row.Url),
		CanReview:    false,
	}
}

func mapIntroVideoItems(rows []queries.TblTeacherIntroVideo) []frontend.IntroVideoItem {
	items := make([]frontend.IntroVideoItem, len(rows))
	for i, row := range rows {
		items[i] = mapIntroVideoItem(row)
	}
	return items
}

func mapAllIntroVideoItems(ctx context.Context, rows []queries.GetAllTeacherIntroVideosFilteredRow) ([]frontend.IntroVideoItem, error) {
	teacherIDs := make([]int64, len(rows))
	items := make([]frontend.IntroVideoItem, len(rows))
	for i, row := range rows {
		teacherIDs[i] = row.TeacherID
		linkLabel := introVideoLinkLabel(row.Url, row.OriginalFilename)
		items[i] = frontend.IntroVideoItem{
			ID:           strconv.FormatInt(row.ID, 10),
			LinkLabel:    linkLabel,
			URL:          nullStringValue(row.Url),
			SourceType:   introVideoSourceType(row.SourceType),
			Status:       constants.TeacherIntroVideoStatus(row.Status),
			UploadedAt:   utils.FormatNullDateTimePHT(row.CreatedAt),
			UploadedBy:   row.TeacherName,
			RejectReason: nullStringValue(row.RejectReason),
			UploadedByAvatar: buildTeacherListAvatarProps(
				row.TeacherID,
				row.TeacherFirstName,
				row.TeacherMiddleName,
				row.TeacherLastName,
				row.TeacherAssignedColor,
				row.TeacherProfilePicture,
			),
			ViewURL:   introVideoViewURL(row.ID, row.Url),
			CanReview: row.Status == string(constants.TeacherIntroVideoStatusSubmitted),
		}
	}

	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, err
	}
	enrichIntroVideoItemsWithRoleBadges(items, teacherIDs, rolesMap)
	return items, nil
}

func handleIntroVideos(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)

	data := frontend.IntroVideosPageData{}

	switch role {
	case auth.RoleSuperuser, auth.RoleAdmin:
		data.Title = "Intro Videos"
		data.Description = "Review teacher introduction video submissions."
		data.ShowUploader = true
		data.ShowActions = true
		data.ShowTeacherFilter = true
	case auth.RoleTeacher:
		data.Title = "My Intro Video"
		data.Description = "Your submitted introduction video."
	default:
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && !constants.ValidTeacherIntroVideoStatus(status) {
		status = ""
	}
	data.Status = constants.TeacherIntroVideoStatus(status)

	writeHTML(w)
	if err := frontend.IntroVideosPage(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleIntroVideosPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)
	filters, err := parseIntroVideoFilters(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sort := parseListSort(r, frontend.ListSortKindIntroVideo)

	var (
		items        []frontend.IntroVideoItem
		showUploader bool
		showActions  bool
		isTeacher    bool
	)

	switch role {
	case auth.RoleSuperuser, auth.RoleAdmin:
		showUploader = true
		showActions = true
		rows, err := dbRO.GetQueries().GetAllTeacherIntroVideosFiltered(ctx, introVideoAllFilterParams(filters))
		if err != nil {
			logs.Log().Error("get filtered teacher intro videos", zap.Error(err))
			HttpError(w, "Failed to load intro videos", http.StatusInternalServerError)
			return
		}
		sortIntroVideoRows(rows, sort)
		items, err = mapAllIntroVideoItems(ctx, rows)
		if err != nil {
			logs.Log().Error("load teacher roles for intro videos", zap.Error(err))
			HttpError(w, "Failed to load intro videos", http.StatusInternalServerError)
			return
		}
	case auth.RoleTeacher:
		isTeacher = true
		rows, err := dbRO.GetQueries().GetTeacherIntroVideosByTeacherIDFiltered(ctx, introVideoTeacherFilterParams(user.ID, filters))
		if err != nil {
			logs.Log().Error("get filtered teacher intro videos", zap.Error(err))
			HttpError(w, "Failed to load intro videos", http.StatusInternalServerError)
			return
		}
		sortTeacherIntroVideoRows(rows, sort)
		items = mapIntroVideoItems(rows)
	default:
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	writeHTML(w)
	if err := frontend.IntroVideosTableBody(items, showUploader, showActions, introVideosEmptyMessage(filters, isTeacher)).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

type introVideoFilters struct {
	Query     string
	Status    string
	TeacherID int64
}

func (f introVideoFilters) active() bool {
	return f.Query != "" || f.Status != "" || f.TeacherID != 0
}

func parseIntroVideoFilters(r *http.Request) (introVideoFilters, error) {
	filters := introVideoFilters{
		Query:  firstQueryParam(r, "introVideoQ", "q"),
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
	}
	if filters.Status != "" && !constants.ValidTeacherIntroVideoStatus(filters.Status) {
		return introVideoFilters{}, ErrInvalidIntroVideoStatus
	}
	teacherIDStr := strings.TrimSpace(r.URL.Query().Get("teacherId"))
	if teacherIDStr != "" {
		teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64)
		if err != nil {
			return introVideoFilters{}, ErrInvalidTeacherID
		}
		filters.TeacherID = teacherID
	}
	return filters, nil
}

func introVideoAllFilterParams(filters introVideoFilters) queries.GetAllTeacherIntroVideosFilteredParams {
	query := sql.NullString{String: filters.Query, Valid: true}
	return queries.GetAllTeacherIntroVideosFilteredParams{
		Column1:   filters.Status,
		Status:    filters.Status,
		Column3:   filters.TeacherID,
		TeacherID: filters.TeacherID,
		Column5:   filters.Query,
		Column6:   query,
		Column7:   query,
		Column8:   query,
	}
}

func introVideoTeacherFilterParams(teacherID int64, filters introVideoFilters) queries.GetTeacherIntroVideosByTeacherIDFilteredParams {
	query := sql.NullString{String: filters.Query, Valid: true}
	return queries.GetTeacherIntroVideosByTeacherIDFilteredParams{
		TeacherID: teacherID,
		Column2:   filters.Status,
		Status:    filters.Status,
		Column4:   filters.Query,
		Column5:   query,
		Column6:   query,
	}
}

func introVideosEmptyMessage(filters introVideoFilters, isTeacher bool) string {
	if filters.active() {
		return "No intro videos match your filters."
	}
	if isTeacher {
		return "No intro video submitted yet. Submit one from My Profile."
	}
	return "No intro videos submitted yet."
}

func handleIntroVideosPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "intro-videos", "/file"); ok {
		handleIntroVideoFile(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "intro-videos", "/approve"); ok {
		handleIntroVideoReview(w, r, id, string(constants.TeacherIntroVideoStatusApproved), "approved", "")
		return
	}
	if id, ok := extractPathID(r, "intro-videos", "/reject"); ok {
		handleIntroVideoReject(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "intro-videos", "/delete"); ok {
		handleIntroVideoDelete(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleProfileIntroVideo(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}
	_, uploadAllowed := introVideoUploadAccessForViewer(ctx)
	if !uploadAllowed {
		setErrorFlash(w, "Intro video submissions are currently disabled")
		HttpRedirect(w, r, "/profile")
		return
	}

	blocking, err := dbRO.GetQueries().HasBlockingTeacherIntroVideo(ctx, user.ID)
	if err != nil {
		logs.Log().Error("check blocking teacher intro video", zap.Error(err))
		setErrorFlash(w, "Failed to verify intro video status")
		HttpRedirect(w, r, "/profile")
		return
	}
	if blocking > 0 {
		setErrorFlash(w, "You already have a submitted or approved intro video. You cannot submit again unless it is rejected or deleted.")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form submission")
		HttpRedirect(w, r, "/profile")
		return
	}

	sourceTypeStr := strings.TrimSpace(r.FormValue("source_type"))
	if !constants.ValidTeacherIntroVideoSourceType(sourceTypeStr) {
		setErrorFlash(w, "Select Google Drive or YouTube as the video source")
		HttpRedirect(w, r, "/profile")
		return
	}

	parsed, err := teacherintrovideo.ParseURL(constants.TeacherIntroVideoSourceType(sourceTypeStr), r.FormValue("url"))
	if err != nil {
		setErrorFlash(w, introVideoSubmitErrorMessage(err))
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().InsertTeacherIntroVideo(ctx, queries.InsertTeacherIntroVideoParams{
		TeacherID:  user.ID,
		Url:        sql.NullString{String: parsed.URL, Valid: true},
		SourceType: sql.NullString{String: string(parsed.SourceType), Valid: true},
		Status:     string(constants.TeacherIntroVideoStatusSubmitted),
	}); err != nil {
		if database.IsUniqueConstraint(err) {
			setErrorFlash(w, "You already have a submitted or approved intro video. You cannot submit again unless it is rejected or deleted.")
		} else {
			logs.Log().Error("insert teacher intro video", zap.Error(err))
			setErrorFlash(w, "Failed to record intro video")
		}
		HttpRedirect(w, r, "/profile")
		return
	}

	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("submitted intro video link for teacher '%s'", user.Name))
	notifySuperuser(ctx, user, notifications.KindIntroVideoSubmitted,
		fmt.Sprintf("Teacher '%s' submitted an intro video link", user.Name), "")
	setSuccessFlash(w, "Intro video submitted successfully. It will be reviewed by an administrator.")
	HttpRedirect(w, r, "/profile")
}

func introVideoSubmitErrorMessage(err error) string {
	switch {
	case errors.Is(err, teacherintrovideo.ErrURLRequired):
		return "Video URL is required"
	case errors.Is(err, teacherintrovideo.ErrInvalidURL):
		return "Enter a valid link for the selected source"
	case errors.Is(err, teacherintrovideo.ErrUnsupportedSourceType):
		return "Select Google Drive or YouTube as the video source"
	default:
		return "Failed to validate intro video link"
	}
}

func handleIntroVideoFile(w http.ResponseWriter, r *http.Request, videoID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetTeacherIntroVideoByID(ctx, videoID)
	if err != nil {
		HttpError(w, "Intro video not found", http.StatusNotFound)
		return
	}

	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(role) && user.ID != row.TeacherID {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}
	if row.Status == string(constants.TeacherIntroVideoStatusDeleted) && !auth.HasAdminAccess(role) {
		HttpError(w, "Intro video not found", http.StatusNotFound)
		return
	}
	if !row.StoredFilename.Valid || row.StoredFilename.String == "" {
		HttpError(w, "Intro video not found", http.StatusNotFound)
		return
	}

	obj, err := storage.Default().Get(ctx, storage.CategoryIntroVideos, row.StoredFilename.String)
	if err != nil {
		HttpError(w, "Intro video not found", http.StatusNotFound)
		return
	}

	filename := "intro-video"
	if row.OriginalFilename.Valid && row.OriginalFilename.String != "" {
		filename = row.OriginalFilename.String
	}
	mimeType := "video/mp4"
	if row.MimeType.Valid && row.MimeType.String != "" {
		mimeType = row.MimeType.String
	}

	serveStorageObject(w, obj, mimeType, map[string]string{
		"Content-Disposition": fmt.Sprintf("inline; filename=%q", filename),
	})
}

func handleIntroVideoReview(w http.ResponseWriter, r *http.Request, videoID int64, status, actionLabel, rejectReason string) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherIntroVideoByID(ctx, videoID)
	if err != nil {
		setErrorFlash(w, "Intro video not found")
		HttpRedirect(w, r, "/intro-videos")
		return
	}
	if row.Status != string(constants.TeacherIntroVideoStatusSubmitted) {
		setErrorFlash(w, "Only submitted intro videos can be reviewed")
		HttpRedirect(w, r, "/intro-videos")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherIntroVideoStatus(ctx, queries.UpdateTeacherIntroVideoStatusParams{
		Status:       status,
		ReviewedBy:   sql.NullInt64{Int64: user.ID, Valid: true},
		RejectReason: sql.NullString{String: rejectReason, Valid: rejectReason != ""},
		ID:           videoID,
	}); err != nil {
		logs.Log().Error("update teacher intro video status", zap.Error(err), zap.String("status", status))
		setErrorFlash(w, "Failed to update intro video status")
		HttpRedirect(w, r, "/intro-videos")
		return
	}

	linkLabel := introVideoLinkLabel(row.Url, row.OriginalFilename)
	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("%s intro video '%s' (id %d)", actionLabel, linkLabel, videoID))
	teacherName := teacherNameByID(ctx, row.TeacherID)
	notifyTeacher(ctx, row.TeacherID, teacherName, user, notifications.KindIntroVideoReviewed,
		fmt.Sprintf("Your intro video '%s' was %s", linkLabel, strings.ToLower(actionLabel)), "")
	setSuccessFlash(w, fmt.Sprintf("Intro video %s successfully.", actionLabel))
	HttpRedirect(w, r, "/intro-videos")
}

func handleIntroVideoReject(w http.ResponseWriter, r *http.Request, videoID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form submission")
		HttpRedirect(w, r, "/intro-videos")
		return
	}

	rejectReason := strings.TrimSpace(r.FormValue("reject_reason"))
	if rejectReason == "" {
		setErrorFlash(w, "A reject reason is required")
		HttpRedirect(w, r, "/intro-videos")
		return
	}

	handleIntroVideoReview(w, r, videoID, string(constants.TeacherIntroVideoStatusRejected), "rejected", rejectReason)
}

func handleIntroVideoDelete(w http.ResponseWriter, r *http.Request, videoID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherIntroVideoByID(ctx, videoID)
	if err != nil {
		setErrorFlash(w, "Intro video not found")
		HttpRedirect(w, r, "/intro-videos")
		return
	}

	if err := dbRW.GetQueries().SoftDeleteTeacherIntroVideo(ctx, queries.SoftDeleteTeacherIntroVideoParams{
		ReviewedBy: sql.NullInt64{Int64: user.ID, Valid: true},
		ID:         videoID,
	}); err != nil {
		logs.Log().Error("soft delete teacher intro video", zap.Error(err), zap.Int64("video_id", videoID))
		setErrorFlash(w, "Failed to delete intro video")
		HttpRedirect(w, r, "/intro-videos")
		return
	}

	if row.StoredFilename.Valid && row.StoredFilename.String != "" {
		if err := storage.Default().Delete(ctx, storage.CategoryIntroVideos, row.StoredFilename.String); err != nil {
			logs.Log().Error("remove intro video file", zap.Error(err), zap.String("filename", row.StoredFilename.String))
		}
	}

	linkLabel := introVideoLinkLabel(row.Url, row.OriginalFilename)
	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("deleted intro video '%s' (id %d)", linkLabel, videoID))
	teacherName := teacherNameByID(ctx, row.TeacherID)
	notifyTeacher(ctx, row.TeacherID, teacherName, user, notifications.KindIntroVideoDeleted,
		fmt.Sprintf("Your intro video '%s' was deleted by an administrator. You may submit a new one.", linkLabel), "")
	setSuccessFlash(w, "Intro video deleted successfully.")
	HttpRedirect(w, r, "/intro-videos")
}
