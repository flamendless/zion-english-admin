package utils

import (
	"net/http"
	"strings"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type SortParams struct {
	By    string
	Order SortOrder
}

func ParseSortOrder(value string) SortOrder {
	if strings.EqualFold(strings.TrimSpace(value), string(SortOrderAsc)) {
		return SortOrderAsc
	}
	return SortOrderDesc
}

func ResolveSortBy(requested string, allowed []string, defaultBy string) string {
	requested = strings.TrimSpace(requested)
	for _, candidate := range allowed {
		if requested == candidate {
			return requested
		}
	}
	return defaultBy
}

func ParseSortParams(r *http.Request, allowed []string, defaultBy string, defaultOrder SortOrder) SortParams {
	by := ResolveSortBy(r.URL.Query().Get("sortBy"), allowed, defaultBy)
	order := ParseSortOrder(r.URL.Query().Get("sortOrder"))
	if r.URL.Query().Get("sortOrder") == "" {
		order = defaultOrder
	}
	return SortParams{By: by, Order: order}
}

func (s SortParams) QueryValues() map[string]string {
	return map[string]string{
		"sortBy":    s.By,
		"sortOrder": string(s.Order),
	}
}

func (s SortParams) OrderSQL() string {
	return string(s.Order)
}
