package cmd

import (
	"fmt"
	"net/http"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/featureflags"
	"zion-english/internal/utils"
)

func handleFeatureFlags(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleFeatureFlagsGet(w, r)
	case http.MethodPost:
		handleFeatureFlagsUpdate(w, r)
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleFeatureFlagsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	zoomEnvConfigured := meetingSvc != nil && meetingSvc.IsZoomConfigured()
	googleEnvConfigured := calendarSvc != nil && calendarSvc.IsConfigured()

	data := frontend.FeatureFlagsData{
		Zoom: frontend.FeatureFlagIntegrationItem{
			Name:          "Zoom",
			Description:   "Allow teachers to connect Zoom accounts for automatic meeting rooms on scheduled classes.",
			LogoURL:       utils.URL("/static/zoom-logo.svg"),
			LogoClass:     "integration-logo",
			LogoWidth:     56,
			LogoHeight:    14,
			EnvConfigured: zoomEnvConfigured,
			ConnectionsAllowed: featureflags.IsEnabled(ctx, dbRO, constants.FeatureFlagIntegrationZoom),
			FormFieldName: "zoom_enabled",
		},
		GoogleCalendar: frontend.FeatureFlagIntegrationItem{
			Name:          "Google Calendar",
			Description:   "Allow teachers to connect Google Calendar so scheduled classes sync to their Zion English calendar.",
			LogoURL:       utils.URL("/static/google-calendar-logo.svg"),
			LogoClass:     "integration-logo integration-logo-calendar",
			LogoWidth:     24,
			LogoHeight:    24,
			EnvConfigured: googleEnvConfigured,
			ConnectionsAllowed: featureflags.IsEnabled(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar),
			FormFieldName: "google_calendar_enabled",
		},
	}

	if err := frontend.FeatureFlags(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleFeatureFlagsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		HttpError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	zoomEnabled := r.FormValue("zoom_enabled") == "on"
	googleEnabled := r.FormValue("google_calendar_enabled") == "on"

	prevZoom := featureflags.IsEnabled(ctx, dbRO, constants.FeatureFlagIntegrationZoom)
	prevGoogle := featureflags.IsEnabled(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar)

	if err := featureflags.SetEnabled(ctx, dbRW, constants.FeatureFlagIntegrationZoom, zoomEnabled); err != nil {
		HttpError(w, fmt.Sprintf("Failed to update Zoom flag: %v", err), http.StatusInternalServerError)
		return
	}
	if err := featureflags.SetEnabled(ctx, dbRW, constants.FeatureFlagIntegrationGoogleCalendar, googleEnabled); err != nil {
		HttpError(w, fmt.Sprintf("Failed to update Google Calendar flag: %v", err), http.StatusInternalServerError)
		return
	}

	user := auth.GetUser(ctx)
	if prevZoom != zoomEnabled {
		if zoomEnabled {
			insertAuditLogAs(ctx, user, "feature-flags", "enabled zoom integration connections")
		} else {
			insertAuditLogAs(ctx, user, "feature-flags", "disabled zoom integration connections")
		}
	}
	if prevGoogle != googleEnabled {
		if googleEnabled {
			insertAuditLogAs(ctx, user, "feature-flags", "enabled google calendar integration connections")
		} else {
			insertAuditLogAs(ctx, user, "feature-flags", "disabled google calendar integration connections")
		}
	}

	setSuccessFlash(w, "Feature flags saved successfully")
	HttpRedirect(w, r, "/feature-flags")
}
