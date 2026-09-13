package cmd

import (
	"net/http"
	"strings"
)

func firstQueryParam(r *http.Request, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(r.URL.Query().Get(key)); v != "" {
			return v
		}
	}
	return ""
}

func parseListDateRange(r *http.Request) (string, string, error) {
	startDate := strings.TrimSpace(r.URL.Query().Get("startDate"))
	endDate := strings.TrimSpace(r.URL.Query().Get("endDate"))
	if startDate != "" && endDate != "" {
		if startDate > endDate {
			return "", "", ErrEndDateBeforeStart
		}
		return startDate, endDate, nil
	}

	preset := strings.TrimSpace(r.URL.Query().Get("datePreset"))
	if preset == "" {
		return "", "", ErrMissingDateRange
	}
	parts := strings.Split(preset, "|")
	if len(parts) != 2 {
		return "", "", ErrInvalidDateRange
	}
	if parts[0] > parts[1] {
		return "", "", ErrEndDateBeforeStart
	}
	return parts[0], parts[1], nil
}
