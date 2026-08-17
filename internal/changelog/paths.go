package changelog

import (
	"os"
	"strings"
)

const (
	EnvChangelogPath = "CHANGELOG_PATH"
	DefaultPath      = "data/changelogs.yaml"
)

func ResolvePath() string {
	if path := strings.TrimSpace(os.Getenv(EnvChangelogPath)); path != "" {
		return path
	}
	return DefaultPath
}
