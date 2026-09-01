package cmd

import (
	"net/http"
	"zion-english/frontend"
	"zion-english/internal/utils"
)

func parseListSort(r *http.Request, kind frontend.ListSortKind) utils.SortParams {
	defaultBy, defaultOrder := frontend.DefaultSortFor(kind)
	allowed := make([]string, 0, len(frontend.SortOptionsFor(kind)))
	for _, opt := range frontend.SortOptionsFor(kind) {
		allowed = append(allowed, opt.Value)
	}
	return utils.ParseSortParams(r, allowed, defaultBy, defaultOrder)
}

func listQueryParamsWithSort(r *http.Request, kind frontend.ListSortKind) map[string]string {
	params := listQueryParams(r)
	for k, v := range parseListSort(r, kind).QueryValues() {
		if v != "" {
			params[k] = v
		}
	}
	return params
}
