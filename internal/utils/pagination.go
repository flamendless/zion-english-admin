package utils

import (
	"net/http"
	"strconv"
)

const (
	DefaultPageSize = 25
	MaxPageSize     = 100
)

type Page struct {
	Number int
	Size   int
	Total  int64
}

func ParsePageQuery(r *http.Request) Page {
	page := Page{Number: 1, Size: DefaultPageSize}

	if n, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && n > 0 {
		page.Number = n
	}
	if s, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil && s > 0 {
		page.Size = s
		if page.Size > MaxPageSize {
			page.Size = MaxPageSize
		}
	}

	return page
}

func (p Page) Offset() int {
	return (p.Number - 1) * p.Size
}

func (p Page) TotalPages() int {
	if p.Total == 0 {
		return 1
	}
	tp := int((p.Total + int64(p.Size) - 1) / int64(p.Size))
	if tp < 1 {
		return 1
	}
	return tp
}

func (p Page) HasPrev() bool {
	return p.Number > 1
}

func (p Page) HasNext() bool {
	return p.Number < p.TotalPages()
}

func QueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func QueryParamInt64(r *http.Request, key string) int64 {
	v, _ := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
	return v
}

func BuildPageURL(basePath string, page Page, params map[string]string) string {
	q := "page=" + strconv.Itoa(page.Number)
	if page.Size > 0 && page.Size != DefaultPageSize {
		q += "&pageSize=" + strconv.Itoa(page.Size)
	}
	for k, v := range params {
		if v != "" {
			q += "&" + k + "=" + v
		}
	}
	if basePath == "" {
		return "?" + q
	}
	sep := "?"
	if len(basePath) > 0 && basePath[len(basePath)-1] == '?' {
		sep = ""
	} else if len(basePath) > 0 && basePath[len(basePath)-1] != '?' && containsQuery(basePath) {
		sep = "&"
	}
	return basePath + sep + q
}

func BuildPageURLAt(basePath string, pageNum int, pageSize int, params map[string]string) string {
	page := Page{Number: pageNum, Size: pageSize}
	return BuildPageURL(basePath, page, params)
}

func containsQuery(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '?' {
			return true
		}
	}
	return false
}
