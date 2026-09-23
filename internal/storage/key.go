package storage

import (
	"path/filepath"
	"strings"
)

func ObjectKey(category Category, filename string) string {
	base := filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	return string(category) + "/" + base
}

func SanitizeFilename(filename string) string {
	return filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
}
