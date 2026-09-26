package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/notifications"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/storage"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

const maxDocumentBytes = 5 << 20

func mapDocumentItems(rows []queries.TblTeacherDocument) []frontend.DocumentItem {
	items := make([]frontend.DocumentItem, len(rows))
	for i, row := range rows {
		items[i] = frontend.DocumentItem{
			ID:         strconv.FormatInt(row.ID, 10),
			Filename:   row.OriginalFilename,
			Extension:  row.FileExtension,
			Type:       row.Type,
			FileSize:   utils.FormatFileSize(row.FileSize),
			Status:     constants.TeacherDocumentStatus(row.Status),
			UploadedAt: utils.FormatNullDateTimePHT(row.UploadedAt),
			UploadedBy: "",
			ViewURL:    utils.URL(fmt.Sprintf("/documents/%d/file", row.ID)),
			CanReview:  false,
		}
	}
	return items
}

func mapAllDocumentItems(ctx context.Context, rows []queries.GetAllTeacherDocumentsFilteredRow) ([]frontend.DocumentItem, error) {
	teacherIDs := make([]int64, len(rows))
	items := make([]frontend.DocumentItem, len(rows))
	for i, row := range rows {
		teacherIDs[i] = row.TeacherID
		items[i] = frontend.DocumentItem{
			ID:         strconv.FormatInt(row.ID, 10),
			Filename:   row.OriginalFilename,
			Extension:  row.FileExtension,
			Type:       row.Type,
			FileSize:   utils.FormatFileSize(row.FileSize),
			Status:     constants.TeacherDocumentStatus(row.Status),
			UploadedAt: utils.FormatNullDateTimePHT(row.UploadedAt),
			UploadedBy: row.TeacherName,
			UploadedByAvatar: buildTeacherListAvatarProps(
				row.TeacherID,
				row.TeacherFirstName,
				row.TeacherMiddleName,
				row.TeacherLastName,
				row.TeacherAssignedColor,
				row.TeacherProfilePicture,
			),
			ViewURL:    utils.URL(fmt.Sprintf("/documents/%d/file", row.ID)),
			CanReview:  row.Type != string(constants.TeacherDocumentTypeResume) && row.Status == string(constants.TeacherDocumentStatusSubmitted),
		}
	}

	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return nil, err
	}
	enrichDocumentItemsWithRoleBadges(items, teacherIDs, rolesMap)
	return items, nil
}

func handleDocuments(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)

	data := frontend.DocumentsPageData{}

	switch role {
	case auth.RoleSuperuser, auth.RoleAdmin:
		data.Title = "Documents"
		data.Description = "All teacher uploads for review."
		data.ShowUploader = true
		data.ShowActions = true
		data.ShowTeacherFilter = true
	case auth.RoleTeacher:
		data.Title = "My Documents"
		data.Description = "Your uploaded profile photos and ID documents."
	default:
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && !constants.ValidTeacherDocumentStatus(status) {
		status = ""
	}
	data.Status = constants.TeacherDocumentStatus(status)

	writeHTML(w)
	if err := frontend.DocumentsPage(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDocumentsPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)
	filters, err := parseDocumentFilters(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sort := parseListSort(r, frontend.ListSortKindDocument)

	var (
		items        []frontend.DocumentItem
		showUploader bool
		showActions  bool
		isTeacher    bool
	)

	switch role {
	case auth.RoleSuperuser, auth.RoleAdmin:
		showUploader = true
		showActions = true
		rows, err := dbRO.GetQueries().GetAllTeacherDocumentsFiltered(ctx, documentAllFilterParams(filters))
		if err != nil {
			logs.Log().Error("get filtered teacher documents", zap.Error(err))
			HttpError(w, "Failed to load documents", http.StatusInternalServerError)
			return
		}
		sortDocumentRows(rows, sort)
		items, err = mapAllDocumentItems(ctx, rows)
		if err != nil {
			logs.Log().Error("load teacher roles for documents", zap.Error(err))
			HttpError(w, "Failed to load documents", http.StatusInternalServerError)
			return
		}
	case auth.RoleTeacher:
		isTeacher = true
		rows, err := dbRO.GetQueries().GetTeacherDocumentsByTeacherIDFiltered(ctx, documentTeacherFilterParams(user.ID, filters))
		if err != nil {
			logs.Log().Error("get filtered teacher documents", zap.Error(err))
			HttpError(w, "Failed to load documents", http.StatusInternalServerError)
			return
		}
		sortTeacherDocumentRows(rows, sort)
		items = mapDocumentItems(rows)
	default:
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	writeHTML(w)
	if err := frontend.DocumentsTableBody(items, showUploader, showActions, documentsEmptyMessage(filters, isTeacher)).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

type documentFilters struct {
	Query     string
	Type      string
	Status    string
	TeacherID int64
}

func (f documentFilters) active() bool {
	return f.Query != "" || f.Type != "" || f.Status != "" || f.TeacherID != 0
}

func parseDocumentFilters(r *http.Request) (documentFilters, error) {
	filters := documentFilters{
		Query:  firstQueryParam(r, "documentQ", "q"),
		Type:   strings.TrimSpace(r.URL.Query().Get("type")),
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
	}
	if filters.Type != "" && !constants.ValidTeacherDocumentType(filters.Type) {
		return documentFilters{}, ErrInvalidDocumentType
	}
	if filters.Status != "" && !constants.ValidTeacherDocumentStatus(filters.Status) {
		return documentFilters{}, ErrInvalidDocumentStatus
	}
	teacherIDStr := strings.TrimSpace(r.URL.Query().Get("teacherId"))
	if teacherIDStr != "" {
		teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64)
		if err != nil {
			return documentFilters{}, ErrInvalidTeacherID
		}
		filters.TeacherID = teacherID
	}
	return filters, nil
}

func documentAllFilterParams(filters documentFilters) queries.GetAllTeacherDocumentsFilteredParams {
	teacherID := filters.TeacherID
	query := filters.Query
	return queries.GetAllTeacherDocumentsFilteredParams{
		Column1:   filters.Type,
		Type:      filters.Type,
		Column3:   filters.Status,
		Status:    filters.Status,
		Column5:   teacherID,
		TeacherID: teacherID,
		Column7:   query,
		Column8:   sql.NullString{String: query, Valid: true},
		Column9:   sql.NullString{String: query, Valid: true},
	}
}

func documentTeacherFilterParams(teacherID int64, filters documentFilters) queries.GetTeacherDocumentsByTeacherIDFilteredParams {
	query := filters.Query
	return queries.GetTeacherDocumentsByTeacherIDFilteredParams{
		TeacherID: teacherID,
		Column2:   filters.Type,
		Type:      filters.Type,
		Column4:   filters.Status,
		Status:    filters.Status,
		Column6:   query,
		Column7:   sql.NullString{String: query, Valid: true},
	}
}

func documentsEmptyMessage(filters documentFilters, isTeacher bool) string {
	if filters.active() {
		return "No documents match your filters."
	}
	if isTeacher {
		return "No documents uploaded yet. Upload your profile photo and valid ID from My Profile."
	}
	return "No documents uploaded yet."
}

func handleDocumentsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "documents", "/file"); ok {
		handleDocumentFile(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "documents", "/approve"); ok {
		handleDocumentReview(w, r, id, string(constants.TeacherDocumentStatusApproved), "approved")
		return
	}
	if id, ok := extractPathID(r, "documents", "/reject"); ok {
		handleDocumentReview(w, r, id, string(constants.TeacherDocumentStatusRejected), "rejected")
		return
	}
	if id, ok := extractPathID(r, "documents", "/delete"); ok {
		handleDocumentDelete(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleProfileDocument(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	blocking, err := dbRO.GetQueries().HasBlockingTeacherDocument(ctx, queries.HasBlockingTeacherDocumentParams{
		TeacherID: user.ID,
		Type:      string(constants.TeacherDocumentTypeDocument),
	})
	if err != nil {
		logs.Log().Error("check blocking teacher document", zap.Error(err))
		setErrorFlash(w, "Failed to verify document status")
		HttpRedirect(w, r, "/profile")
		return
	}
	if blocking > 0 {
		setErrorFlash(w, "You already have a submitted or approved ID document. You cannot upload again unless it is rejected.")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := r.ParseMultipartForm(maxDocumentBytes); err != nil {
		setErrorFlash(w, "File is too large. Maximum size is 5 MB.")
		HttpRedirect(w, r, "/profile")
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		setErrorFlash(w, "Please choose a document to upload")
		HttpRedirect(w, r, "/profile")
		return
	}
	defer file.Close()

	ext, err := validateDocumentUpload(file, header.Filename, header.Size)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/profile")
		return
	}

	storedFilename := fmt.Sprintf("%d_%d%s", user.ID, time.Now().UnixNano(), ext)
	storageKey := storage.TeacherDocumentKey(constants.TeacherDocumentTypeDocument, storedFilename)
	store := storage.Default()
	if err := store.Put(ctx, storage.CategoryTeacherDocuments, storageKey, file, documentContentType(strings.TrimPrefix(ext, "."))); err != nil {
		logs.Log().Error("write document file", zap.Error(err))
		insertUploadLog(ctx, user, uploadLogEntry{
			Module:        "profile",
			Outcome:       constants.UploadLogOutcomeFailed,
			Kind:          constants.UploadLogKindDocument,
			Summary:       fmt.Sprintf("ID document storage failed for teacher '%s' (id %d), file '%s': %v", user.Name, user.ID, filepath.Base(header.Filename), err),
			Filename:      filepath.Base(header.Filename),
			FileSize:      header.Size,
			FileSizeValid: true,
		})
		setErrorFlash(w, "Failed to save document")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().InsertTeacherDocument(ctx, queries.InsertTeacherDocumentParams{
		TeacherID:        user.ID,
		Type:             string(constants.TeacherDocumentTypeDocument),
		OriginalFilename: filepath.Base(header.Filename),
		StoredFilename:   storedFilename,
		FileExtension:    strings.TrimPrefix(ext, "."),
		FileSize:         header.Size,
		Status:           string(constants.TeacherDocumentStatusSubmitted),
	}); err != nil {
		_ = store.Delete(ctx, storage.CategoryTeacherDocuments, storageKey)
		logs.Log().Error("insert teacher document", zap.Error(err))
		insertUploadLog(ctx, user, uploadLogEntry{
			Module:        "profile",
			Outcome:       constants.UploadLogOutcomeFailed,
			Kind:          constants.UploadLogKindDocument,
			Summary:       fmt.Sprintf("ID document database insert failed for teacher '%s' (id %d), file '%s': %v", user.Name, user.ID, filepath.Base(header.Filename), err),
			Filename:      filepath.Base(header.Filename),
			FileSize:      header.Size,
			FileSizeValid: true,
		})
		setErrorFlash(w, "Failed to record document")
		HttpRedirect(w, r, "/profile")
		return
	}

	insertUploadLog(ctx, user, uploadLogEntry{
		Module:        "profile",
		Outcome:       constants.UploadLogOutcomeSucceeded,
		Kind:          constants.UploadLogKindDocument,
		Summary:       fmt.Sprintf("ID document for teacher '%s' (id %d), file '%s'", user.Name, user.ID, filepath.Base(header.Filename)),
		Filename:      filepath.Base(header.Filename),
		FileSize:      header.Size,
		FileSizeValid: true,
	})
	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("submitted ID document for teacher '%s'", user.Name))
	notifySuperuser(ctx, user, notifications.KindDocumentSubmitted,
		fmt.Sprintf("Teacher '%s' submitted ID document '%s'", user.Name, filepath.Base(header.Filename)), "")
	setSuccessFlash(w, "ID document submitted successfully. It will be reviewed by an administrator.")
	HttpRedirect(w, r, "/profile")
}

func handleProfileResume(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	lastUploaded, err := dbRO.GetQueries().GetLatestTeacherDocumentUploadedAtByTeacherIDAndType(ctx, queries.GetLatestTeacherDocumentUploadedAtByTeacherIDAndTypeParams{
		TeacherID: user.ID,
		Type:      string(constants.TeacherDocumentTypeResume),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logs.Log().Error("check latest teacher resume upload", zap.Error(err))
		setErrorFlash(w, "Failed to verify resume upload status")
		HttpRedirect(w, r, "/profile")
		return
	}
	if allowed, days := utils.ResumeUploadAllowed(lastUploaded, time.Now()); !allowed {
		setErrorFlash(w, fmt.Sprintf("You can upload a new resume/CV again in %d day(s).", days))
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := r.ParseMultipartForm(maxDocumentBytes); err != nil {
		setErrorFlash(w, "File is too large. Maximum size is 5 MB.")
		HttpRedirect(w, r, "/profile")
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		setErrorFlash(w, "Please choose a resume/CV to upload")
		HttpRedirect(w, r, "/profile")
		return
	}
	defer file.Close()

	ext, err := validateDocumentUpload(file, header.Filename, header.Size)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/profile")
		return
	}

	storedFilename := fmt.Sprintf("%d_%d%s", user.ID, time.Now().UnixNano(), ext)
	storageKey := storage.TeacherDocumentKey(constants.TeacherDocumentTypeResume, storedFilename)
	store := storage.Default()
	if err := store.Put(ctx, storage.CategoryTeacherDocuments, storageKey, file, documentContentType(strings.TrimPrefix(ext, "."))); err != nil {
		logs.Log().Error("write resume file", zap.Error(err))
		insertUploadLog(ctx, user, uploadLogEntry{
			Module:        "profile",
			Outcome:       constants.UploadLogOutcomeFailed,
			Kind:          constants.UploadLogKindDocument,
			Summary:       fmt.Sprintf("resume/CV storage failed for teacher '%s' (id %d), file '%s': %v", user.Name, user.ID, filepath.Base(header.Filename), err),
			Filename:      filepath.Base(header.Filename),
			FileSize:      header.Size,
			FileSizeValid: true,
		})
		setErrorFlash(w, "Failed to save resume/CV")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().InsertTeacherDocument(ctx, queries.InsertTeacherDocumentParams{
		TeacherID:        user.ID,
		Type:             string(constants.TeacherDocumentTypeResume),
		OriginalFilename: filepath.Base(header.Filename),
		StoredFilename:   storedFilename,
		FileExtension:    strings.TrimPrefix(ext, "."),
		FileSize:         header.Size,
		Status:           string(constants.TeacherDocumentStatusApproved),
	}); err != nil {
		_ = store.Delete(ctx, storage.CategoryTeacherDocuments, storageKey)
		logs.Log().Error("insert teacher resume", zap.Error(err))
		insertUploadLog(ctx, user, uploadLogEntry{
			Module:        "profile",
			Outcome:       constants.UploadLogOutcomeFailed,
			Kind:          constants.UploadLogKindDocument,
			Summary:       fmt.Sprintf("resume/CV database insert failed for teacher '%s' (id %d), file '%s': %v", user.Name, user.ID, filepath.Base(header.Filename), err),
			Filename:      filepath.Base(header.Filename),
			FileSize:      header.Size,
			FileSizeValid: true,
		})
		setErrorFlash(w, "Failed to record resume/CV")
		HttpRedirect(w, r, "/profile")
		return
	}

	insertUploadLog(ctx, user, uploadLogEntry{
		Module:        "profile",
		Outcome:       constants.UploadLogOutcomeSucceeded,
		Kind:          constants.UploadLogKindDocument,
		Summary:       fmt.Sprintf("resume/CV for teacher '%s' (id %d), file '%s'", user.Name, user.ID, filepath.Base(header.Filename)),
		Filename:      filepath.Base(header.Filename),
		FileSize:      header.Size,
		FileSizeValid: true,
	})
	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("uploaded resume/CV for teacher '%s'", user.Name))
	setSuccessFlash(w, "Resume/CV uploaded successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleDocumentFile(w http.ResponseWriter, r *http.Request, documentID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetTeacherDocumentByID(ctx, documentID)
	if err != nil {
		HttpError(w, "Document not found", http.StatusNotFound)
		return
	}

	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(role) && user.ID != row.TeacherID {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	var obj *storage.Object
	if row.Type == string(constants.TeacherDocumentTypeAvatar) {
		obj, err = storage.Default().Get(ctx, storage.CategoryAvatars, row.StoredFilename)
	} else {
		obj, err = storage.GetTeacherDocument(ctx, storage.Default(), constants.TeacherDocumentType(row.Type), row.StoredFilename)
	}
	if err != nil {
		HttpError(w, "Document not found", http.StatusNotFound)
		return
	}

	serveStorageObject(w, obj, documentContentType(row.FileExtension), map[string]string{
		"Content-Disposition": fmt.Sprintf("inline; filename=%q", row.OriginalFilename),
	})
}

func handleDocumentReview(w http.ResponseWriter, r *http.Request, documentID int64, status, actionLabel string) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherDocumentByID(ctx, documentID)
	if err != nil {
		setErrorFlash(w, "Document not found")
		HttpRedirect(w, r, "/documents")
		return
	}
	if row.Type == string(constants.TeacherDocumentTypeResume) {
		setErrorFlash(w, "Resume/CV documents cannot be reviewed")
		HttpRedirect(w, r, "/documents")
		return
	}
	if row.Status != string(constants.TeacherDocumentStatusSubmitted) {
		setErrorFlash(w, "Only submitted documents can be reviewed")
		HttpRedirect(w, r, "/documents")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherDocumentStatus(ctx, queries.UpdateTeacherDocumentStatusParams{
		Status:     status,
		ReviewedBy: sql.NullInt64{Int64: user.ID, Valid: true},
		ID:         documentID,
	}); err != nil {
		logs.Log().Error("update teacher document status", zap.Error(err), zap.String("status", status))
		setErrorFlash(w, "Failed to update document status")
		HttpRedirect(w, r, "/documents")
		return
	}

	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("%s document '%s' (id %d)", actionLabel, row.OriginalFilename, documentID))
	teacherName := teacherNameByID(ctx, row.TeacherID)
	notifyTeacher(ctx, row.TeacherID, teacherName, user, notifications.KindDocumentReviewed,
		fmt.Sprintf("Your document '%s' was %s", row.OriginalFilename, strings.ToLower(actionLabel)), "")
	setSuccessFlash(w, fmt.Sprintf("Document %s successfully.", actionLabel))
	HttpRedirect(w, r, "/documents")
}

func handleDocumentDelete(w http.ResponseWriter, r *http.Request, documentID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.HasAdminAccess(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherDocumentByID(ctx, documentID)
	if err != nil {
		setErrorFlash(w, "Document not found")
		HttpRedirect(w, r, "/documents")
		return
	}

	if err := dbRW.GetQueries().DeleteTeacherDocument(ctx, documentID); err != nil {
		logs.Log().Error("delete teacher document", zap.Error(err), zap.Int64("document_id", documentID))
		setErrorFlash(w, "Failed to delete document")
		HttpRedirect(w, r, "/documents")
		return
	}

	if row.Type == string(constants.TeacherDocumentTypeAvatar) {
		profile, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, row.TeacherID)
		if err == nil && profile.ProfilePicture.Valid && profile.ProfilePicture.String == row.StoredFilename {
			if err := dbRW.GetQueries().UpdateTeacherProfilePicture(ctx, queries.UpdateTeacherProfilePictureParams{
				ProfilePicture: sql.NullString{Valid: false},
				ID:             row.TeacherID,
			}); err != nil {
				logs.Log().Error("clear teacher profile picture after document delete", zap.Error(err), zap.Int64("teacher_id", row.TeacherID))
			}
		}
	}

	if row.Type == string(constants.TeacherDocumentTypeAvatar) {
		if err := storage.Default().Delete(ctx, storage.CategoryAvatars, row.StoredFilename); err != nil {
			logs.Log().Error("remove document file", zap.Error(err), zap.String("filename", row.StoredFilename))
		}
	} else if err := storage.DeleteTeacherDocument(ctx, storage.Default(), constants.TeacherDocumentType(row.Type), row.StoredFilename); err != nil {
		logs.Log().Error("remove document file", zap.Error(err), zap.String("filename", row.StoredFilename))
	}

	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("deleted document '%s' (id %d)", row.OriginalFilename, documentID))
	setSuccessFlash(w, "Document deleted successfully.")
	HttpRedirect(w, r, "/documents")
}

func validateDocumentUpload(file io.ReadSeeker, filename string, size int64) (string, error) {
	if size <= 0 {
		return "", ErrUploadedFileEmpty
	}
	if size > maxDocumentBytes {
		return "", ErrDocumentFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg":
		_, format, err := image.DecodeConfig(file)
		if err != nil {
			return "", ErrInvalidDocumentImage
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", ErrFailedToReadUploadedFile
		}
		if ext == ".jpeg" || format == "jpeg" {
			return ".jpg", nil
		}
		return ext, nil
	case ".pdf":
		header := make([]byte, 4)
		if _, err := io.ReadFull(file, header); err != nil {
			return "", ErrInvalidPDFFile
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", ErrFailedToReadUploadedFile
		}
		if !bytes.Equal(header, []byte("%PDF")) {
			return "", ErrInvalidPDFFile
		}
		return ".pdf", nil
	default:
		return "", ErrUnsupportedDocumentFormat
	}
}

func documentContentType(ext string) string {
	switch strings.ToLower(ext) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

func logAvatarDocument(ctx context.Context, teacherID int64, originalFilename, storedFilename, ext string, size int64) error {
	return dbRW.GetQueries().InsertTeacherDocument(ctx, queries.InsertTeacherDocumentParams{
		TeacherID:        teacherID,
		Type:             string(constants.TeacherDocumentTypeAvatar),
		OriginalFilename: originalFilename,
		StoredFilename:   storedFilename,
		FileExtension:    strings.TrimPrefix(ext, "."),
		FileSize:         size,
		Status:           string(constants.TeacherDocumentStatusApproved),
	})
}
