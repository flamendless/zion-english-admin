package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// E2EDBRelativePath is the Playwright E2E SQLite file under the repo root (never data/zion.db).
const E2EDBRelativePath = "e2e/.data/e2e.db"

// E2EDBPath returns the absolute path to the isolated E2E database. It does not read ZION_DB_PATH.
func E2EDBPath() string {
	root := repoRootFromCWD()
	abs := filepath.Clean(filepath.Join(root, E2EDBRelativePath))
	if !strings.HasSuffix(abs, filepath.Join("e2e", ".data", "e2e.db")) {
		panic(fmt.Sprintf("e2e database path must end with e2e/.data/e2e.db, got %q", abs))
	}
	AssertNeverApplicationDB(abs)
	return abs
}

// AssertNeverApplicationDB panics if path is the default local application database.
func AssertNeverApplicationDB(abs string) {
	defaultAbs, err := filepath.Abs(defaultDBPath)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve default db path: %v", err))
	}
	if abs == defaultAbs {
		panic("refused: path is the application database (data/zion.db)")
	}
	if strings.HasSuffix(filepath.Clean(abs), string(os.PathSeparator)+"data"+string(os.PathSeparator)+"zion.db") {
		panic("refused: path is under data/zion.db")
	}
}

func repoRootFromCWD() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

// RemoveSQLiteDBFiles deletes the main DB file and WAL sidecar files.
func RemoveSQLiteDBFiles(path string) {
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		_ = os.Remove(p)
	}
}
