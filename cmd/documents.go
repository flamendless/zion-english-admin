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
	"os"
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
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

const (
	documentDir      = "data/teacher-documents"
	maxDocumentBytes = 5 << 20
)

func ensureDocumentDir() error {
	return os.MkdirAll(documentDir, 0755)
}

func documentFilePath(filename string) string {
	return filepath.Join(documentDir, filepath.Base(filename))
}

func documentStoragePath(row queries.TblTeacherDocument) string {
	if row.Type == string(constants.TeacherDocumentTypeAvatar) {
		return avatarFilePath(row.StoredFilename)
	}
	return documentFilePath(row.StoredFilename)
}

func mapDocumentItems(rows []queries.TblTeacherDocument) []frontend.DocumentItem {
	items := make([]frontend.DocumentItem, len(rows))
	for i, row := range rows {
		items[i] = frontend.DocumentItem{
			ID:         strconv.FormatInt(row.ID, 10),
			Filename:   row.OriginalFilename,
			Extension:  row.FileExtension,
			Type:       row.Type,
			FileSize:   utils.FormatFileSize(row.FileSize),
			Status:     row.Status,
			UploadedAt: utils.FormatNullDateTimePHT(row.UploadedAt),
			UploadedBy: "",
			ViewURL:    utils.URL(fmt.Sprintf("/documents/%d/file", row.ID)),
			CanReview:  false,
		}
	}
	return items
}

func mapAllDocumentItems(rows []queries.GetAllTeacherDocumentsRow) []frontend.DocumentItem {
	items := make([]frontend.DocumentItem, len(rows))
	for i, row := range rows {
		items[i] = frontend.DocumentItem{
			ID:         strconv.FormatInt(row.ID, 10),
			Filename:   row.OriginalFilename,
			Extension:  row.FileExtension,
			Type:       row.Type,
			FileSize:   utils.FormatFileSize(row.FileSize),
			Status:     row.Status,
			UploadedAt: utils.FormatNullDateTimePHT(row.UploadedAt),
			UploadedBy: row.TeacherName,
			ViewURL:    utils.URL(fmt.Sprintf("/documents/%d/file", row.ID)),
			CanReview:  row.Status == string(constants.TeacherDocumentStatusSubmitted),
		}
	}
	return items
}

func handleDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	role := auth.GetRole(ctx)
	user := auth.GetUser(ctx)

	data := frontend.DocumentsPageData{}

	switch role {
	case auth.RoleSuperuser:
		data.Title = "Documents"
		data.Description = "All teacher uploads for review."
		data.ShowUploader = true
		data.ShowActions = true
		rows, err := dbRO.GetQueries().GetAllTeacherDocuments(ctx)
		if err != nil {
			logs.Log().Error("get all teacher documents", zap.Error(err))
			HttpError(w, "Failed to load documents", http.StatusInternalServerError)
			return
		}
		data.Documents = mapAllDocumentItems(rows)
	case auth.RoleTeacher:
		data.Title = "My Documents"
		data.Description = "Your uploaded profile photos and ID documents."
		rows, err := dbRO.GetQueries().GetTeacherDocumentsByTeacherID(ctx, user.ID)
		if err != nil {
			logs.Log().Error("get teacher documents", zap.Error(err))
			HttpError(w, "Failed to load documents", http.StatusInternalServerError)
			return
		}
		data.Documents = mapDocumentItems(rows)
	default:
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.DocumentsPage(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
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
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleProfileDocument(w http.ResponseWriter, r *http.Request) {
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

	blocking, err := dbRO.GetQueries().HasBlockingTeacherDocument(ctx, user.ID)
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

	if err := ensureDocumentDir(); err != nil {
		logs.Log().Error("create document dir", zap.Error(err))
		setErrorFlash(w, "Failed to prepare upload")
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
	destPath := documentFilePath(storedFilename)

	out, err := os.Create(destPath)
	if err != nil {
		logs.Log().Error("create document file", zap.Error(err))
		setErrorFlash(w, "Failed to save document")
		HttpRedirect(w, r, "/profile")
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		_ = os.Remove(destPath)
		logs.Log().Error("write document file", zap.Error(err))
		setErrorFlash(w, "Failed to save document")
		HttpRedirect(w, r, "/profile")
		return
	}
	out.Close()

	if err := dbRW.GetQueries().InsertTeacherDocument(ctx, queries.InsertTeacherDocumentParams{
		TeacherID:        user.ID,
		Type:             string(constants.TeacherDocumentTypeDocument),
		OriginalFilename: filepath.Base(header.Filename),
		StoredFilename:   storedFilename,
		FileExtension:    strings.TrimPrefix(ext, "."),
		FileSize:         header.Size,
		Status:           string(constants.TeacherDocumentStatusSubmitted),
	}); err != nil {
		_ = os.Remove(destPath)
		logs.Log().Error("insert teacher document", zap.Error(err))
		setErrorFlash(w, "Failed to record document")
		HttpRedirect(w, r, "/profile")
		return
	}

	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("submitted ID document for teacher '%s'", user.Name))
	notifySuperuser(ctx, user, notifications.KindDocumentSubmitted,
		fmt.Sprintf("Teacher '%s' submitted ID document '%s'", user.Name, filepath.Base(header.Filename)), "")
	setSuccessFlash(w, "ID document submitted successfully. It will be reviewed by an administrator.")
	HttpRedirect(w, r, "/profile")
}

func handleDocumentFile(w http.ResponseWriter, r *http.Request, documentID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	if role != auth.RoleSuperuser && user.ID != row.TeacherID {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	path := documentStoragePath(row)
	if _, err := os.Stat(path); err != nil {
		HttpError(w, "Document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", documentContentType(row.FileExtension))
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", row.OriginalFilename))
	http.ServeFile(w, r, path)
}

func handleDocumentReview(w http.ResponseWriter, r *http.Request, documentID int64, status, actionLabel string) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleSuperuser {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherDocumentByID(ctx, documentID)
	if err != nil {
		setErrorFlash(w, "Document not found")
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
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleSuperuser {
		HttpError(w, "Access denied", http.StatusForbidden)
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

	filePath := documentStoragePath(row)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logs.Log().Error("remove document file", zap.Error(err), zap.String("path", filePath))
	}

	insertAuditLogAs(ctx, user, "teachers", fmt.Sprintf("deleted document '%s' (id %d)", row.OriginalFilename, documentID))
	setSuccessFlash(w, "Document deleted successfully.")
	HttpRedirect(w, r, "/documents")
}

func validateDocumentUpload(file io.ReadSeeker, filename string, size int64) (string, error) {
	if size <= 0 {
		return "", errors.New("Uploaded file is empty")
	}
	if size > maxDocumentBytes {
		return "", errors.New("File is too large. Maximum size is 5 MB.")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg":
		_, format, err := image.DecodeConfig(file)
		if err != nil {
			return "", errors.New("Invalid image file. Please upload a PNG or JPEG image.")
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", errors.New("Failed to read uploaded file")
		}
		if ext == ".jpeg" || format == "jpeg" {
			return ".jpg", nil
		}
		return ext, nil
	case ".pdf":
		header := make([]byte, 4)
		if _, err := io.ReadFull(file, header); err != nil {
			return "", errors.New("Invalid PDF file")
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", errors.New("Failed to read uploaded file")
		}
		if !bytes.Equal(header, []byte("%PDF")) {
			return "", errors.New("Invalid PDF file")
		}
		return ".pdf", nil
	default:
		return "", errors.New("Unsupported file format. Please upload PNG, JPEG, JPG, or PDF.")
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
