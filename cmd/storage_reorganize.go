package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/logs"
	"zion-english/internal/storage"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var cmdStorageReorganizeTeacherDocuments = &cobra.Command{
	Use:   "reorganize-teacher-documents",
	Short: "Move teacher document files into type-based subfolders",
	Long: strings.TrimSpace(`
Move legacy flat teacher document files into document/ and resume/ subfolders
under data/teacher-documents. When R2 is configured, also copies objects to typed
keys and removes legacy flat keys. Idempotent and does not change the database.`),
	Run: func(cmd *cobra.Command, args []string) {
		runStorageReorganizeTeacherDocuments()
	},
}

func init() {
	cmdStorage.AddCommand(cmdStorageReorganizeTeacherDocuments)
}

func runStorageReorganizeTeacherDocuments() {
	cfg := conf.Conf()
	if err := storage.Init(cfg); err != nil {
		panic(fmt.Sprintf("failed to initialize storage: %v", err))
	}
	if err := database.Init("data/zion.db"); err != nil {
		panic(fmt.Sprintf("failed to initialize database: %v", err))
	}
	defer database.Close()

	ctx := context.Background()
	dbRO := database.New(database.DB_MODE_RO)
	store := storage.Default()

	rows, err := dbRO.GetQueries().GetTeacherDocumentsForStorageReorganize(ctx)
	if err != nil {
		panic(fmt.Sprintf("load teacher documents: %v", err))
	}

	var movedLocal, movedRemote, skipped int
	for _, row := range rows {
		docType := constants.TeacherDocumentType(row.Type)
		typedKey := storage.TeacherDocumentKey(docType, row.StoredFilename)
		basename := storage.SanitizeFilename(row.StoredFilename)
		if typedKey == basename {
			skipped++
			continue
		}

		didMoveLocal, err := reorganizeLocalTeacherDocument("data/teacher-documents", basename, typedKey)
		if err != nil {
			panic(fmt.Sprintf("reorganize local %s: %v", basename, err))
		}
		if didMoveLocal {
			movedLocal++
			fmt.Printf("moved local: %s -> %s\n", basename, typedKey)
		}

		if cfg.R2Enabled() && store.Backend() == "r2" {
			didMoveRemote, err := reorganizeRemoteTeacherDocument(ctx, store, basename, typedKey)
			if err != nil {
				panic(fmt.Sprintf("reorganize remote %s: %v", basename, err))
			}
			if didMoveRemote {
				movedRemote++
				fmt.Printf("moved remote: %s -> %s\n", basename, typedKey)
			}
		}
	}

	fmt.Printf("Reorganize complete. local_moved=%d remote_moved=%d skipped=%d\n", movedLocal, movedRemote, skipped)
	logs.Log().Info("storage reorganize teacher documents finished",
		zap.Int("local_moved", movedLocal),
		zap.Int("remote_moved", movedRemote),
		zap.Int("skipped", skipped),
	)
}

func reorganizeLocalTeacherDocument(baseDir, basename, typedKey string) (bool, error) {
	flatPath := filepath.Join(baseDir, basename)
	typedPath := filepath.Join(baseDir, typedKey)

	if _, err := os.Stat(flatPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if _, err := os.Stat(typedPath); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(typedPath), 0755); err != nil {
		return false, err
	}
	if err := os.Rename(flatPath, typedPath); err != nil {
		return false, err
	}
	return true, nil
}

func reorganizeRemoteTeacherDocument(ctx context.Context, store storage.Storage, basename, typedKey string) (bool, error) {
	existsFlat, err := store.Exists(ctx, storage.CategoryTeacherDocuments, basename)
	if err != nil {
		return false, err
	}
	if !existsFlat {
		return false, nil
	}
	existsTyped, err := store.Exists(ctx, storage.CategoryTeacherDocuments, typedKey)
	if err != nil {
		return false, err
	}
	if existsTyped {
		return false, nil
	}

	obj, err := store.Get(ctx, storage.CategoryTeacherDocuments, basename)
	if err != nil {
		return false, err
	}
	defer obj.Reader.Close()

	body, err := io.ReadAll(obj.Reader)
	if err != nil {
		return false, err
	}

	contentType := obj.ContentType
	if contentType == "" {
		contentType = contentTypeForPath(basename)
	}
	if err := store.Put(ctx, storage.CategoryTeacherDocuments, typedKey, bytes.NewReader(body), contentType); err != nil {
		return false, err
	}
	if err := store.Delete(ctx, storage.CategoryTeacherDocuments, basename); err != nil {
		return false, err
	}
	return true, nil
}
