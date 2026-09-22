package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/notifications"
	"zion-english/internal/utils"

	"go.uber.org/zap"
	"gopkg.in/vansante/go-ffprobe.v2"
)

const introVideoDir = "data/teacher-intro-videos"

var allowedIntroVideoExtensions = map[string]string{
	".mp4":  "video/mp4",
	".m4v":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
	".ogv":  "video/ogg",
	".3gp":  "video/3gpp",
}

func ensureIntroVideoDir() error {
	return os.MkdirAll(introVideoDir, 0755)
}

func introVideoFilePath(filename string) string {
	return filepath.Join(introVideoDir, filepath.Base(filename))
}

func mapIntroVideoItems(rows []queries.TblTeacherIntroVideo) []frontend.IntroVideoItem {
	items := make([]frontend.IntroVideoItem, len(rows))
	for i, row := range rows {
		items[i] = frontend.IntroVideoItem{
			ID:           strconv.FormatInt(row.ID, 10),
			Filename:     row.OriginalFilename,
			MimeType:     row.MimeType,
			FileSize:     utils.FormatFileSize(row.FileSize),
			Status:       constants.TeacherIntroVideoStatus(row.Status),
			UploadedAt:   utils.FormatNullDateTimePHT(row.CreatedAt),
			RejectReason: nullStringValue(row.RejectReason),
			ViewURL:      utils.URL(fmt.Sprintf("/intro-videos/%d/file", row.ID)),
			CanReview:    false,
		}
	}
	return items
}

func mapAllIntroVideoItems(ctx context.Context, rows []queries.GetAllTeacherIntroVideosFilteredRow) ([]frontend.IntroVideoItem, error) {
	teacherIDs := make([]int64, len(rows))
	items := make([]frontend.IntroVideoItem, len(rows))
	for i, row := range rows {
		teacherIDs[i] = row.TeacherID
		items[i] = frontend.IntroVideoItem{
			ID:           strconv.FormatInt(row.ID, 10),
			Filename:     row.OriginalFilename,
			MimeType:     row.MimeType,
			FileSize:     utils.FormatFileSize(row.FileSize),
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
			ViewURL:   utils.URL(fmt.Sprintf("/intro-videos/%d/file", row.ID)),
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
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)

	data := frontend.IntroVideosPageData{}

	switch role {
	case auth.RoleSuperuser, auth.RoleAdmin:
		data.Title = "Intro Videos"
		data.Description = "Review teacher introduction video uploads."
		data.ShowUploader = true
		data.ShowActions = true
		data.ShowTeacherFilter = true
	case auth.RoleTeacher:
		data.Title = "My Intro Video"
		data.Description = "Your submitted introduction video."
	default:
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && !constants.ValidTeacherIntroVideoStatus(status) {
		status = ""
	}
	data.Status = constants.TeacherIntroVideoStatus(status)

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.IntroVideosPage(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleIntroVideosPartial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html")
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
	return queries.GetAllTeacherIntroVideosFilteredParams{
		Column1:   filters.Status,
		Status:    filters.Status,
		Column3:   filters.TeacherID,
		TeacherID: filters.TeacherID,
		Column5:   filters.Query,
		Column6:   sql.NullString{String: filters.Query, Valid: true},
		Column7:   sql.NullString{String: filters.Query, Valid: true},
	}
}

func introVideoTeacherFilterParams(teacherID int64, filters introVideoFilters) queries.GetTeacherIntroVideosByTeacherIDFilteredParams {
	return queries.GetTeacherIntroVideosByTeacherIDFilteredParams{
		TeacherID: teacherID,
		Column2:   filters.Status,
		Status:    filters.Status,
		Column4:   filters.Query,
		Column5:   sql.NullString{String: filters.Query, Valid: true},
	}
}

func introVideosEmptyMessage(filters introVideoFilters, isTeacher bool) string {
	if filters.active() {
		return "No intro videos match your filters."
	}
	if isTeacher {
		return "No intro video uploaded yet. Upload one from My Profile."
	}
	return "No intro videos uploaded yet."
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
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleProfileIntroVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}
	_, uploadAllowed := introVideoUploadAccessForViewer(ctx)
	if !uploadAllowed {
		setErrorFlash(w, "Intro video uploads are currently disabled")
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
		setErrorFlash(w, "You already have a submitted or approved intro video. You cannot upload again unless it is rejected or deleted.")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := ensureIntroVideoDir(); err != nil {
		logs.Log().Error("create intro video dir", zap.Error(err))
		setErrorFlash(w, "Failed to prepare upload")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := r.ParseMultipartForm(constants.MaxIntroVideoBytes); err != nil {
		setErrorFlash(w, "File is too large. Maximum size is 20 MB.")
		HttpRedirect(w, r, "/profile")
		return
	}

	file, header, err := r.FormFile("intro_video")
	if err != nil {
		setErrorFlash(w, "Please choose a video to upload")
		HttpRedirect(w, r, "/profile")
		return
	}
	defer file.Close()

	tempPath, mimeType, err := saveTempIntroVideo(file, header.Filename, header.Size)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/profile")
		return
	}
	defer os.Remove(tempPath)

	duration, err := probeVideoDuration(ctx, tempPath)
	if err != nil {
		if errors.Is(err, ErrFfprobeUnavailable) {
			setErrorFlash(w, "Video duration check requires ffprobe. Install ffmpeg and ensure ffprobe is on PATH.")
		} else {
			setErrorFlash(w, "Could not read video duration. Please upload a valid video file.")
		}
		HttpRedirect(w, r, "/profile")
		return
	}
	if duration > constants.MaxIntroVideoDurationSeconds {
		setErrorFlash(w, fmt.Sprintf("Video is too long. Maximum duration is %d seconds.", constants.MaxIntroVideoDurationSeconds))
		HttpRedirect(w, r, "/profile")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	storedFilename := fmt.Sprintf("%d_%d%s", user.ID, time.Now().UnixNano(), ext)
	destPath := introVideoFilePath(storedFilename)

	if err := os.Rename(tempPath, destPath); err != nil {
		if copyErr := copyFile(tempPath, destPath); copyErr != nil {
			logs.Log().Error("move intro video file", zap.Error(err), zap.Error(copyErr))
			setErrorFlash(w, "Failed to save video")
			HttpRedirect(w, r, "/profile")
			return
		}
	}

	if err := dbRW.GetQueries().InsertTeacherIntroVideo(ctx, queries.InsertTeacherIntroVideoParams{
		TeacherID:        user.ID,
		OriginalFilename: filepath.Base(header.Filename),
		StoredFilename:   storedFilename,
		MimeType:         mimeType,
		FileSize:         header.Size,
		Status:           string(constants.TeacherIntroVideoStatusSubmitted),
	}); err != nil {
		_ = os.Remove(destPath)
		if database.IsUniqueConstraint(err) {
			setErrorFlash(w, "You already have a submitted or approved intro video. You cannot upload again unless it is rejected or deleted.")
		} else {
			logs.Log().Error("insert teacher intro video", zap.Error(err))
			setErrorFlash(w, "Failed to record intro video")
		}
		HttpRedirect(w, r, "/profile")
		return
	}

	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("submitted intro video for teacher '%s'", user.Name))
	notifySuperuser(ctx, user, notifications.KindIntroVideoSubmitted,
		fmt.Sprintf("Teacher '%s' submitted intro video '%s'", user.Name, filepath.Base(header.Filename)), "")
	setSuccessFlash(w, "Intro video submitted successfully. It will be reviewed by an administrator.")
	HttpRedirect(w, r, "/profile")
}

func handleIntroVideoFile(w http.ResponseWriter, r *http.Request, videoID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}
	if row.Status == string(constants.TeacherIntroVideoStatusDeleted) && !auth.HasAdminAccess(role) {
		HttpError(w, "Intro video not found", http.StatusNotFound)
		return
	}

	path := introVideoFilePath(row.StoredFilename)
	if _, err := os.Stat(path); err != nil {
		HttpError(w, "Intro video not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", row.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", row.OriginalFilename))
	http.ServeFile(w, r, path)
}

func handleIntroVideoReview(w http.ResponseWriter, r *http.Request, videoID int64, status, actionLabel, rejectReason string) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(auth.GetRole(ctx)) {
		HttpError(w, "Access denied", http.StatusForbidden)
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

	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("%s intro video '%s' (id %d)", actionLabel, row.OriginalFilename, videoID))
	teacherName := teacherNameByID(ctx, row.TeacherID)
	notifyTeacher(ctx, row.TeacherID, teacherName, user, notifications.KindIntroVideoReviewed,
		fmt.Sprintf("Your intro video '%s' was %s", row.OriginalFilename, strings.ToLower(actionLabel)), "")
	setSuccessFlash(w, fmt.Sprintf("Intro video %s successfully.", actionLabel))
	HttpRedirect(w, r, "/intro-videos")
}

func handleIntroVideoReject(w http.ResponseWriter, r *http.Request, videoID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(auth.GetRole(ctx)) {
		HttpError(w, "Access denied", http.StatusForbidden)
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

	filePath := introVideoFilePath(row.StoredFilename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logs.Log().Error("remove intro video file", zap.Error(err), zap.String("path", filePath))
	}

	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("deleted intro video '%s' (id %d)", row.OriginalFilename, videoID))
	teacherName := teacherNameByID(ctx, row.TeacherID)
	notifyTeacher(ctx, row.TeacherID, teacherName, user, notifications.KindIntroVideoDeleted,
		fmt.Sprintf("Your intro video '%s' was deleted by an administrator. You may upload a new one.", row.OriginalFilename), "")
	setSuccessFlash(w, "Intro video deleted successfully.")
	HttpRedirect(w, r, "/intro-videos")
}

func saveTempIntroVideo(file io.ReadSeeker, filename string, size int64) (string, string, error) {
	mimeType, err := validateIntroVideoUpload(file, filename, size)
	if err != nil {
		return "", "", err
	}

	ext := strings.ToLower(filepath.Ext(filename))
	tempFile, err := os.CreateTemp(introVideoDir, "upload-*"+ext)
	if err != nil {
		return "", "", ErrIntroVideoUploadPrepareFailed
	}
	tempPath := tempFile.Name()

	if _, err := io.Copy(tempFile, file); err != nil {
		tempFile.Close()
		_ = os.Remove(tempPath)
		return "", "", ErrIntroVideoSaveFailed
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", "", ErrIntroVideoSaveFailed
	}

	return tempPath, mimeType, nil
}

func validateIntroVideoUpload(file io.ReadSeeker, filename string, size int64) (string, error) {
	if size <= 0 {
		return "", ErrIntroVideoFileEmpty
	}
	if size > constants.MaxIntroVideoBytes {
		return "", ErrIntroVideoFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(filename))
	mimeType, ok := allowedIntroVideoExtensions[ext]
	if !ok {
		return "", ErrUnsupportedIntroVideoFormat
	}

	header := make([]byte, 12)
	if _, err := io.ReadFull(file, header); err != nil {
		return "", ErrInvalidIntroVideoFile
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", ErrIntroVideoReadFailed
	}
	if !looksLikeVideo(header, ext) {
		return "", ErrInvalidIntroVideoContent
	}

	return mimeType, nil
}

func looksLikeVideo(header []byte, ext string) bool {
	switch ext {
	case ".mp4", ".m4v", ".mov":
		return len(header) >= 8 && bytes.Equal(header[4:8], []byte("ftyp"))
	case ".webm", ".mkv":
		return len(header) >= 4 && header[0] == 0x1A && header[1] == 0x45 && header[2] == 0xDF && header[3] == 0xA3
	case ".avi":
		return len(header) >= 12 && bytes.Equal(header[0:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("AVI "))
	case ".ogv":
		return len(header) >= 4 && bytes.Equal(header[0:4], []byte("OggS"))
	case ".3gp":
		return len(header) >= 8 && bytes.Equal(header[4:8], []byte("ftyp"))
	default:
		return false
	}
}

func probeVideoDuration(ctx context.Context, path string) (float64, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	data, err := ffprobe.ProbeURL(probeCtx, path)
	if err != nil {
		if _, ok := errors.AsType[*exec.Error](err); ok {
			return 0, ErrFfprobeUnavailable
		}
		return 0, err
	}
	if data.Format == nil || data.Format.DurationSeconds <= 0 {
		return 0, ErrEmptyIntroVideoDuration
	}
	return data.Format.DurationSeconds, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
