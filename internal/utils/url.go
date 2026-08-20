package utils

import (
	"strings"

	"zion-english/internal/conf"
)

func URL(path string) string {
	base := strings.TrimSuffix(conf.Conf().BasePath, "/")
	if base == "" || base == "/" {
		return path
	}
	return base + path
}

func BasePath() string {
	return conf.Conf().BasePath
}
