package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"zion-english/internal/conf"
	"zion-english/internal/constants"
)

func TestObjectKey(t *testing.T) {
	got := ObjectKey(CategoryAvatars, "1.jpg")
	if got != "avatars/1.jpg" {
		t.Fatalf("got %q want avatars/1.jpg", got)
	}
	got = ObjectKey(CategoryTeacherDocuments, "../evil.pdf")
	if got != "teacher-documents/evil.pdf" {
		t.Fatalf("got %q want teacher-documents/evil.pdf", got)
	}
	got = ObjectKey(CategoryTeacherDocuments, "document/foo.pdf")
	if got != "teacher-documents/document/foo.pdf" {
		t.Fatalf("got %q want teacher-documents/document/foo.pdf", got)
	}
}

func TestSanitizeRelativePath(t *testing.T) {
	if got := SanitizeRelativePath("../evil"); got != "" {
		t.Fatalf("got %q want empty for traversal", got)
	}
	if got := SanitizeRelativePath("document/foo.pdf"); got != "document/foo.pdf" {
		t.Fatalf("got %q want document/foo.pdf", got)
	}
	if got := SanitizeRelativePath("a/b/c.pdf"); got != "" {
		t.Fatalf("got %q want empty for deep path", got)
	}
}

func TestTeacherDocumentKey(t *testing.T) {
	got := TeacherDocumentKey(constants.TeacherDocumentTypeResume, "42_1.pdf")
	if got != "resume/42_1.pdf" {
		t.Fatalf("got %q want resume/42_1.pdf", got)
	}
}

func TestLocalStorageTypedTeacherDocumentRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := &LocalStorage{root: dir}
	ctx := context.Background()
	key := TeacherDocumentKey(constants.TeacherDocumentTypeDocument, "42_1.pdf")
	body := []byte("document bytes")

	if err := store.Put(ctx, CategoryTeacherDocuments, key, bytes.NewReader(body), "application/pdf"); err != nil {
		t.Fatalf("put: %v", err)
	}
	obj, err := store.Get(ctx, CategoryTeacherDocuments, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got, err := io.ReadAll(obj.Reader)
	obj.Reader.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("got %q want %q", got, body)
	}
}

func TestGetTeacherDocumentLegacyFallback(t *testing.T) {
	dir := t.TempDir()
	store := &LocalStorage{root: dir}
	ctx := context.Background()
	body := []byte("legacy bytes")

	if err := store.Put(ctx, CategoryTeacherDocuments, "42_1.pdf", bytes.NewReader(body), "application/pdf"); err != nil {
		t.Fatalf("put legacy: %v", err)
	}
	obj, err := GetTeacherDocument(ctx, store, constants.TeacherDocumentTypeDocument, "42_1.pdf")
	if err != nil {
		t.Fatalf("get teacher document: %v", err)
	}
	got, err := io.ReadAll(obj.Reader)
	obj.Reader.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("got %q want %q", got, body)
	}
}

func TestGetTeacherDocumentTypedPreferredOverLegacy(t *testing.T) {
	dir := t.TempDir()
	store := &LocalStorage{root: dir}
	ctx := context.Background()

	if err := store.Put(ctx, CategoryTeacherDocuments, "42_1.pdf", bytes.NewReader([]byte("legacy")), "application/pdf"); err != nil {
		t.Fatalf("put legacy: %v", err)
	}
	typedKey := TeacherDocumentKey(constants.TeacherDocumentTypeDocument, "42_1.pdf")
	if err := store.Put(ctx, CategoryTeacherDocuments, typedKey, bytes.NewReader([]byte("typed")), "application/pdf"); err != nil {
		t.Fatalf("put typed: %v", err)
	}
	obj, err := GetTeacherDocument(ctx, store, constants.TeacherDocumentTypeDocument, "42_1.pdf")
	if err != nil {
		t.Fatalf("get teacher document: %v", err)
	}
	got, err := io.ReadAll(obj.Reader)
	obj.Reader.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, []byte("typed")) {
		t.Fatalf("got %q want typed path content", got)
	}
}

func TestGetTeacherDocumentMissing(t *testing.T) {
	dir := t.TempDir()
	store := &LocalStorage{root: dir}
	ctx := context.Background()

	_, err := GetTeacherDocument(ctx, store, constants.TeacherDocumentTypeDocument, "missing.pdf")
	if !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("got %v want ErrObjectNotFound", err)
	}
}

func TestLocalStorageRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := &LocalStorage{root: dir}
	ctx := context.Background()

	body := []byte("avatar bytes")
	if err := store.Put(ctx, CategoryAvatars, "42.jpg", bytes.NewReader(body), "image/jpeg"); err != nil {
		t.Fatalf("put: %v", err)
	}
	exists, err := store.Exists(ctx, CategoryAvatars, "42.jpg")
	if err != nil || !exists {
		t.Fatalf("exists: %v %v", exists, err)
	}
	obj, err := store.Get(ctx, CategoryAvatars, "42.jpg")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got, err := io.ReadAll(obj.Reader)
	obj.Reader.Close()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("got %q want %q", got, body)
	}
	if err := store.Delete(ctx, CategoryAvatars, "42.jpg"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	exists, err = store.Exists(ctx, CategoryAvatars, "42.jpg")
	if err != nil || exists {
		t.Fatalf("exists after delete: %v %v", exists, err)
	}
}

func TestConfigR2Enabled(t *testing.T) {
	cfg := &conf.Config{
		Storage: conf.StorageConfig{
			R2: conf.R2Config{
				AccountID:       "acct",
				Bucket:          "bucket",
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
			},
		},
	}
	if !cfg.R2Enabled() {
		t.Fatal("expected R2 enabled")
	}
	cfg.Storage.R2.AccessKeyID = ""
	if cfg.R2Enabled() {
		t.Fatal("expected R2 disabled without access key")
	}
}

func TestContentLengthForPut(t *testing.T) {
	body := []byte("teacher document bytes")
	length, ok := contentLengthForPut(bytes.NewReader(body))
	if !ok || length != int64(len(body)) {
		t.Fatalf("got length=%d ok=%v want %d true", length, ok, len(body))
	}
}

func TestNewFromConfigLocalFallback(t *testing.T) {
	cfg := &conf.Config{AppEnv: conf.EnvLocal}
	store, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("new from config: %v", err)
	}
	if store.Backend() != "local" {
		t.Fatalf("got backend %q want local", store.Backend())
	}
}
