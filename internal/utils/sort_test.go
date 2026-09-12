package utils

import (
	"net/http"
	"testing"
)

func TestParseSortOrder(t *testing.T) {
	tests := []struct {
		input string
		want  SortOrder
	}{
		{"asc", SortOrderAsc},
		{"ASC", SortOrderAsc},
		{"desc", SortOrderDesc},
		{"", SortOrderDesc},
		{"invalid", SortOrderDesc},
	}
	for _, tt := range tests {
		if got := ParseSortOrder(tt.input); got != tt.want {
			t.Errorf("ParseSortOrder(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveSortBy(t *testing.T) {
	allowed := []string{"name", "created_at", "status"}
	if got := ResolveSortBy("name", allowed, "created_at"); got != "name" {
		t.Fatalf("expected name, got %q", got)
	}
	if got := ResolveSortBy("invalid", allowed, "created_at"); got != "created_at" {
		t.Fatalf("expected default created_at, got %q", got)
	}
}

func TestParseSortParams(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/?sortBy=name&sortOrder=asc", nil)
	params := ParseSortParams(req, []string{"name", "created_at"}, "created_at", SortOrderDesc)
	if params.By != "name" || params.Order != SortOrderAsc {
		t.Fatalf("unexpected params: %+v", params)
	}

	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	params = ParseSortParams(req, []string{"name", "created_at"}, "created_at", SortOrderDesc)
	if params.By != "created_at" || params.Order != SortOrderDesc {
		t.Fatalf("unexpected default params: %+v", params)
	}
}
