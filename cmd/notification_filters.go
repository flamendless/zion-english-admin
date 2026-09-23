package cmd

import (
	"net/http"
	"sort"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type notificationFilters struct {
	Message string
	From    string
	Date    string
}

func parseNotificationFilters(r *http.Request) notificationFilters {
	return notificationFilters{
		Message: strings.TrimSpace(r.URL.Query().Get("message")),
		From:    strings.TrimSpace(r.URL.Query().Get("from")),
		Date:    utils.NormalizeDatePHT(r.URL.Query().Get("date")),
	}
}

func (f notificationFilters) active() bool {
	return f.Message != "" || f.From != "" || f.Date != ""
}

func notificationFilterParams(unreadOnly bool, filters notificationFilters, sort utils.SortParams) map[string]string {
	params := make(map[string]string)
	if unreadOnly {
		params["filter"] = "unread"
	}
	if filters.Message != "" {
		params["message"] = filters.Message
	}
	if filters.From != "" {
		params["from"] = filters.From
	}
	if filters.Date != "" {
		params["date"] = filters.Date
	}
	for k, v := range sort.QueryValues() {
		if v != "" {
			params[k] = v
		}
	}
	return params
}

func notificationFromOptions(rows []queries.TblNotification) []frontend.StatusOption {
	seen := make(map[string]bool)
	names := make([]string, 0)
	for _, row := range rows {
		name := strings.TrimSpace(row.FromName)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)
	opts := make([]frontend.StatusOption, 0, len(names))
	for _, name := range names {
		opts = append(opts, frontend.StatusOption{Value: name, Label: name})
	}
	return opts
}

func filterNotificationRows(rows []queries.TblNotification, filters notificationFilters) []queries.TblNotification {
	if !filters.active() {
		return rows
	}
	messageQuery := strings.ToLower(filters.Message)
	filtered := make([]queries.TblNotification, 0, len(rows))
	for _, row := range rows {
		if messageQuery != "" && !strings.Contains(strings.ToLower(row.Message), messageQuery) {
			continue
		}
		if filters.From != "" && row.FromName != filters.From {
			continue
		}
		if filters.Date != "" {
			rowDate, err := notificationCreatedDatePHT(row.CreatedAt)
			if err != nil || rowDate != filters.Date {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func notificationCreatedDatePHT(value string) (string, error) {
	t, err := parseNotificationCreatedAt(value)
	if err != nil {
		return "", err
	}
	return utils.DatePHT(t), nil
}

func parseNotificationCreatedAt(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, ErrEmptyNotificationCreatedAt
	}
	layouts := []string{constants.DateTimeSecondsLayout, constants.DateTimeLayout}
	for _, layout := range layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, ErrInvalidNotificationCreatedAt
}
