package metatags

import (
	"context"
	"net/http"
	"strings"

	"zion-english/internal/database/queries"
)

type contextKey string

const tagsKey contextKey = "head_meta_tags"

type Tag struct {
	Name    string
	Content string
	Value   string
}

func Get(ctx context.Context) []Tag {
	tags, _ := ctx.Value(tagsKey).([]Tag)
	return tags
}

func Middleware(db *queries.Queries, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if !shouldLoadTags(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		rows, err := db.GetAllMetaTags(ctx)
		if err == nil {
			tags := make([]Tag, 0, len(rows))
			for _, row := range rows {
				tags = append(tags, Tag{
					Name:    row.Name,
					Content: row.Content,
					Value:   row.Value,
				})
			}
			ctx = context.WithValue(ctx, tagsKey, tags)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func shouldLoadTags(path string) bool {
	if strings.Contains(path, "/static/") {
		return false
	}
	if strings.HasSuffix(path, "/health") {
		return false
	}
	if strings.Contains(path, "/api/") {
		return false
	}
	if strings.Contains(path, "/partials/") {
		return false
	}
	if strings.Contains(path, "/unread-count") {
		return false
	}
	if strings.Contains(path, "/notifications/panel") {
		return false
	}
	return true
}
