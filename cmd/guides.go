package cmd

import (
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/conf"
	"zion-english/internal/logs"

	"go.uber.org/zap"
)

func handleGuides(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := frontend.GuidesData{
		Guides: []frontend.GuideEntry{
			{
				Slug:        "getting-started",
				Title:       "Getting Started",
				Description: "Learn the basics: profile, students, classes, schedule, and editing records.",
				Audience:    "For teachers",
			},
		},
	}

	if err := frontend.Guides(data).Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render guides page", zap.Error(err))
	}
}

func handleGuidesPath(w http.ResponseWriter, r *http.Request) {
	slug, ok := extractGuideSlug(r)
	if !ok {
		HttpError(w, "Not found", http.StatusNotFound)
		return
	}

	switch slug {
	case "getting-started":
		handleGuideGettingStarted(w, r)
	default:
		HttpError(w, "Not found", http.StatusNotFound)
	}
}

func handleGuideGettingStarted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := frontend.GuideGettingStarted().Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render getting started guide", zap.Error(err))
	}
}

func extractGuideSlug(r *http.Request) (string, bool) {
	cfg := conf.Conf()
	prefix := strings.TrimSuffix(cfg.BasePath, "/") + "/guides/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		return "", false
	}
	slug := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if slug == "" || strings.Contains(slug, "/") {
		return "", false
	}
	return slug, true
}
