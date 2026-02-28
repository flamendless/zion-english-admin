package utils

import (
	"zion-english/internal/conf"
)

func URL(path string) string {
	if conf.Conf().IsProd() {
		return path
	}
	return "/zion-english-admin" + path
}
