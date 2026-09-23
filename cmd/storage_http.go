package cmd

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/storage"
)

func serveStorageObject(w http.ResponseWriter, obj *storage.Object, contentType string, extraHeaders map[string]string) {
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else if obj.ContentType != "" {
		w.Header().Set("Content-Type", obj.ContentType)
	}
	for key, value := range extraHeaders {
		w.Header().Set(key, value)
	}
	_, _ = io.Copy(w, obj.Reader)
	obj.Reader.Close()
}

func documentStorageCategory(row queries.TblTeacherDocument) storage.Category {
	if row.Type == string(constants.TeacherDocumentTypeAvatar) {
		return storage.CategoryAvatars
	}
	return storage.CategoryTeacherDocuments
}

func avatarContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func reportOutputBasename(outputPath string) string {
	if outputPath == "" {
		return ""
	}
	cleanPath := filepath.Clean(strings.ReplaceAll(outputPath, "\\", "/"))
	base := filepath.Base(cleanPath)
	if base == "" || base == "." {
		return ""
	}
	return base
}

func reportCacheAvailable(ctx context.Context, outputPath string) (string, bool) {
	base := reportOutputBasename(outputPath)
	if base == "" {
		return "", false
	}
	store := storage.Default()
	exists, err := store.Exists(ctx, storage.CategoryReports, base)
	if err == nil && exists {
		return base, true
	}
	legacyPath := filepath.Join("tmp", base)
	if _, err := os.Stat(legacyPath); err == nil {
		return base, true
	}
	return "", false
}

func serveReportDownload(w http.ResponseWriter, r *http.Request, filename string) bool {
	base := reportOutputBasename(filename)
	if base == "" {
		return false
	}
	store := storage.Default()
	obj, err := store.Get(r.Context(), storage.CategoryReports, base)
	if err == nil {
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+base+"\"")
		serveStorageObject(w, obj, "", nil)
		return true
	}
	if !errors.Is(err, storage.ErrObjectNotFound) {
		return false
	}
	legacyPath := filepath.Join("tmp", base)
	if _, err := os.Stat(legacyPath); err != nil {
		return false
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+base+"\"")
	http.ServeFile(w, r, legacyPath)
	return true
}

func putReportFile(ctx context.Context, filename string, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return storage.Default().Put(ctx, storage.CategoryReports, filename, f, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
}
