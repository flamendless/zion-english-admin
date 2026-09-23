package storage

import (
	"bytes"
	"context"
	"io"
	"testing"

	"zion-english/internal/conf"
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
