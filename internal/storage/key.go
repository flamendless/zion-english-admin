package storage

import (
	"path/filepath"
	"strings"
)

func ObjectKey(category Category, filename string) string {
	rel := SanitizeRelativePath(filename)
	if rel == "" {
		rel = SanitizeFilename(filename)
	}
	return string(category) + "/" + rel
}

func SanitizeFilename(filename string) string {
	return filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
}

func SanitizeRelativePath(path string) string {
	path = strings.Trim(strings.ReplaceAll(path, "\\", "/"), "/")
	if path == "" || strings.Contains(path, "..") {
		return ""
	}
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		base := filepath.Base(parts[0])
		if base == "" || base == "." {
			return ""
		}
		return base
	}
	if len(parts) != 2 {
		return ""
	}
	subdir := filepath.Base(parts[0])
	base := filepath.Base(parts[1])
	if subdir == "" || subdir == "." || base == "" || base == "." {
		return ""
	}
	return subdir + "/" + base
}
