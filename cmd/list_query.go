package cmd

import (
	"errors"
	"net/http"
	"strings"
)

func parseListDateRange(r *http.Request) (string, string, error) {
	startDate := strings.TrimSpace(r.URL.Query().Get("startDate"))
	endDate := strings.TrimSpace(r.URL.Query().Get("endDate"))
	if startDate != "" && endDate != "" {
		if startDate > endDate {
			return "", "", errors.New("end date must be after start date")
		}
		return startDate, endDate, nil
	}

	preset := strings.TrimSpace(r.URL.Query().Get("datePreset"))
	if preset == "" {
		return "", "", errors.New("missing date range")
	}
	parts := strings.Split(preset, "|")
	if len(parts) != 2 {
		return "", "", errors.New("invalid date range")
	}
	if parts[0] > parts[1] {
		return "", "", errors.New("end date must be after start date")
	}
	return parts[0], parts[1], nil
}
