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
				Access:      "teacher",
			},
			{
				Slug:        "connect-zoom",
				Title:       "Connect Zoom",
				Description: "Link your Zoom account so scheduled classes get a meeting room automatically.",
				Access:      "teacher",
			},
			{
				Slug:        "reports-and-generation",
				Title:       "Reports & Generation",
				Description: "Review teacher payroll summaries, preview class records, and export XLSX reports.",
				Access:      "superuser",
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
	case "connect-zoom":
		handleGuideConnectZoom(w, r)
	case "reports-and-generation":
		handleGuideReportsGeneration(w, r)
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

func handleGuideConnectZoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := frontend.GuideConnectZoom().Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render connect zoom guide", zap.Error(err))
	}
}

func handleGuideReportsGeneration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := frontend.GuideReportsGeneration().Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render reports and generation guide", zap.Error(err))
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
