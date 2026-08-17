package cmd

import (
	"net/http"
	"strconv"
	"strings"
	"zion-english/internal/conf"
)

func extractPathID(r *http.Request, segment, suffix string) (int64, bool) {
	cfg := conf.Conf()
	prefix := strings.TrimSuffix(cfg.BasePath, "/") + "/" + segment + "/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		return 0, false
	}
	rest := strings.TrimPrefix(r.URL.Path, prefix)
	if suffix != "" {
		if !strings.HasSuffix(rest, suffix) {
			return 0, false
		}
		rest = strings.TrimSuffix(rest, suffix)
	}
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func listQueryParams(r *http.Request) map[string]string {
	return map[string]string{
		"q":         r.URL.Query().Get("q"),
		"status":    r.URL.Query().Get("status"),
		"teacherId": r.URL.Query().Get("teacherId"),
		"email":     r.URL.Query().Get("email"),
		"module":    r.URL.Query().Get("module"),
		"startDate": r.URL.Query().Get("startDate"),
		"endDate":   r.URL.Query().Get("endDate"),
	}
}
