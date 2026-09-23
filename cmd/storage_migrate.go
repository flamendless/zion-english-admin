package cmd

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"zion-english/internal/conf"
	"zion-english/internal/logs"
	"zion-english/internal/storage"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var cmdStorage = &cobra.Command{
	Use:   "storage",
	Short: "Object storage utilities",
}

var cmdStorageMigrate = &cobra.Command{
	Use:   "migrate",
	Short: "Upload existing local files to Cloudflare R2",
	Long: strings.TrimSpace(`
Upload files from data/avatars, data/teacher-documents, data/teacher-intro-videos,
and report XLSX files in tmp/.

Requires R2_ACCOUNT_ID, R2_BUCKET, R2_ACCESS_KEY_ID, and R2_SECRET_ACCESS_KEY in .env.
Local files are not deleted after upload.`),
	Run: func(cmd *cobra.Command, args []string) {
		runStorageMigrate()
	},
}

func init() {
	cmdStorage.AddCommand(cmdStorageMigrate)
	rootCmd.AddCommand(cmdStorage)
}

func runStorageMigrate() {
	cfg := conf.Conf()
	if !cfg.R2Enabled() {
		panic("R2 is not configured: set R2_ACCOUNT_ID, R2_BUCKET, R2_ACCESS_KEY_ID, and R2_SECRET_ACCESS_KEY in .env")
	}
	if err := storage.Init(cfg); err != nil {
		panic(fmt.Sprintf("failed to initialize storage: %v", err))
	}

	store := storage.Default()
	if store.Backend() != "r2" {
		panic("storage backend is not r2")
	}

	ctx := context.Background()
	logs.Log().Info("storage migrate starting",
		zap.String("bucket", cfg.Storage.R2.Bucket),
	)

	totalUploaded := 0
	totalSkipped := 0

	for _, job := range []struct {
		dir      string
		category storage.Category
	}{
		{"data/avatars", storage.CategoryAvatars},
		{"data/teacher-documents", storage.CategoryTeacherDocuments},
		{"data/teacher-intro-videos", storage.CategoryIntroVideos},
	} {
		uploaded, skipped, err := migrateLocalDir(ctx, store, job.dir, job.category)
		if err != nil {
			panic(fmt.Sprintf("migrate %s: %v", job.dir, err))
		}
		totalUploaded += uploaded
		totalSkipped += skipped
		fmt.Printf("%s: uploaded=%d skipped=%d\n", job.category, uploaded, skipped)
	}

	reportUploaded, reportSkipped, err := migrateReportFiles(ctx, store)
	if err != nil {
		panic(fmt.Sprintf("migrate reports: %v", err))
	}
	totalUploaded += reportUploaded
	totalSkipped += reportSkipped
	fmt.Printf("%s: uploaded=%d skipped=%d\n", storage.CategoryReports, reportUploaded, reportSkipped)

	fmt.Println("Migration complete. Local files were not deleted.")
	logs.Log().Info("storage migrate finished",
		zap.Int("uploaded", totalUploaded),
		zap.Int("skipped", totalSkipped),
	)
}

func migrateLocalDir(ctx context.Context, store storage.Storage, dir string, category storage.Category) (int, int, error) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}

	var uploaded, skipped int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(dir, name)
		didUpload, err := migrateLocalFile(ctx, store, path, category, name)
		if err != nil {
			return uploaded, skipped, err
		}
		if didUpload {
			uploaded++
			fmt.Printf("upload: %s/%s\n", category, name)
		} else {
			skipped++
			fmt.Printf("skip (exists): %s/%s\n", category, name)
		}
	}
	return uploaded, skipped, nil
}

func migrateLocalFile(ctx context.Context, store storage.Storage, path string, category storage.Category, name string) (bool, error) {
	exists, err := store.Exists(ctx, category, name)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if err := store.Put(ctx, category, name, f, contentTypeForPath(path)); err != nil {
		return false, err
	}
	return true, nil
}

func migrateReportFiles(ctx context.Context, store storage.Storage) (int, int, error) {
	dir := "tmp"
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}

	var uploaded, skipped int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.EqualFold(filepath.Ext(name), ".xlsx") {
			continue
		}
		path := filepath.Join(dir, name)
		didUpload, err := migrateLocalFile(ctx, store, path, storage.CategoryReports, name)
		if err != nil {
			return uploaded, skipped, err
		}
		if didUpload {
			uploaded++
			fmt.Printf("upload: %s/%s\n", storage.CategoryReports, name)
		} else {
			skipped++
			fmt.Printf("skip (exists): %s/%s\n", storage.CategoryReports, name)
		}
	}
	return uploaded, skipped, nil
}

func contentTypeForPath(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return "application/octet-stream"
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	return "application/octet-stream"
}
