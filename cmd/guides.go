package cmd

import (
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
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
				Slug:        constants.GuideSlugGettingStarted,
				Title:       "Getting Started",
				Description: "Learn the basics: profile, students, classes, schedule, and editing records.",
				Access:      "teacher",
			},
			{
				Slug:        constants.GuideSlugConnectZoom,
				Title:       "Connect Zoom",
				Description: "Link your Zoom account so scheduled classes get a meeting room automatically.",
				Access:      "teacher",
			},
			{
				Slug:        constants.GuideSlugConnectGoogleCalendar,
				Title:       "Connect Google Calendar",
				Description: "Link Google Calendar so scheduled classes are added to your Zion English calendar automatically.",
				Access:      "teacher",
			},
			{
				Slug:        constants.GuideSlugFAQ,
				Title:       "FAQ",
				Description: "Answers about scheduled vs recorded classes, absences, earnings, and more.",
				Access:      "teacher",
			},
			{
				Slug:        constants.GuideSlugReportsAndGeneration,
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

	switch constants.GuideSlug(slug) {
	case constants.GuideSlugGettingStarted:
		handleGuideGettingStarted(w, r)
	case constants.GuideSlugConnectZoom:
		handleGuideConnectZoom(w, r)
	case constants.GuideSlugConnectGoogleCalendar:
		handleGuideConnectGoogleCalendar(w, r)
	case constants.GuideSlugFAQ:
		handleGuideFAQ(w, r)
	case constants.GuideSlugReportsAndGeneration:
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

func handleGuideConnectGoogleCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := frontend.GuideConnectGoogleCalendar().Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render connect google calendar guide", zap.Error(err))
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

func handleGuideFAQ(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := frontend.GuideFAQ().Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render faq guide", zap.Error(err))
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
