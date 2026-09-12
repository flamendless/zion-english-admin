package cmd

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/featureflags"
	"zion-english/internal/utils"
)

func classOverdueGracePeriodMinutes(ctx context.Context) int64 {
	return featureflags.ClassOverdueGracePeriodMinutes(ctx, dbRO)
}

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

	zoomEnabled, zoomVisibleRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationZoom)
	googleEnabled, googleVisibleRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar)
	roleOptions := constants.AllTeacherRoles()

	data := frontend.FeatureFlagsData{
		ClassOverdueGracePeriodMinutes: classOverdueGracePeriodMinutes(ctx),
		Zoom: frontend.FeatureFlagIntegrationItem{
			Name:               "Zoom",
			Description:        "Allow teachers to connect Zoom accounts for automatic meeting rooms on scheduled classes.",
			LogoURL:            utils.URL("/static/zoom-logo.svg"),
			LogoClass:          "integration-logo",
			LogoWidth:          56,
			LogoHeight:         14,
			EnvConfigured:      zoomEnvConfigured,
			ConnectionsAllowed: zoomEnabled,
			VisibleRoles:       zoomVisibleRoles,
			RoleOptions:        roleOptions,
			FormFieldName:      "zoom_enabled",
			FormPrefix:         "zoom",
		},
		GoogleCalendar: frontend.FeatureFlagIntegrationItem{
			Name:               "Google Calendar",
			Description:        "Allow teachers to connect Google Calendar so scheduled classes sync to their Zion English calendar.",
			LogoURL:            utils.URL("/static/google-calendar-logo.svg"),
			LogoClass:          "integration-logo integration-logo-calendar",
			LogoWidth:          24,
			LogoHeight:         24,
			EnvConfigured:      googleEnvConfigured,
			ConnectionsAllowed: googleEnabled,
			VisibleRoles:       googleVisibleRoles,
			RoleOptions:        roleOptions,
			FormFieldName:      "google_calendar_enabled",
			FormPrefix:         "google_calendar",
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
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/feature-flags")
		return
	}

	zoomEnabled := r.FormValue("zoom_enabled") == "on"
	googleEnabled := r.FormValue("google_calendar_enabled") == "on"

	zoomRoles, err := parseFeatureFlagRolesFromForm(r, "zoom")
	if err != nil {
		setErrorFlash(w, "Select at least one role for Zoom visibility")
		HttpRedirect(w, r, "/feature-flags")
		return
	}
	googleRoles, err := parseFeatureFlagRolesFromForm(r, "google_calendar")
	if err != nil {
		setErrorFlash(w, "Select at least one role for Google Calendar visibility")
		HttpRedirect(w, r, "/feature-flags")
		return
	}

	prevZoomEnabled, prevZoomRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationZoom)
	prevGoogleEnabled, prevGoogleRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar)
	prevGracePeriod := classOverdueGracePeriodMinutes(ctx)

	gracePeriod, err := parseClassOverdueGracePeriodFromForm(r)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/feature-flags")
		return
	}

	if err := featureflags.SetFlag(ctx, dbRW, constants.FeatureFlagIntegrationZoom, zoomEnabled, zoomRoles); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update Zoom flag: %v", err))
		HttpRedirect(w, r, "/feature-flags")
		return
	}
	if err := featureflags.SetFlag(ctx, dbRW, constants.FeatureFlagIntegrationGoogleCalendar, googleEnabled, googleRoles); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update Google Calendar flag: %v", err))
		HttpRedirect(w, r, "/feature-flags")
		return
	}
	if err := featureflags.SetIntValue(ctx, dbRW, constants.FeatureFlagClassOverdueGracePeriod, gracePeriod); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update class overdue grace period: %v", err))
		HttpRedirect(w, r, "/feature-flags")
		return
	}

	user := auth.GetUser(ctx)
	if prevZoomEnabled != zoomEnabled {
		if zoomEnabled {
			insertAuditLogAs(ctx, user, "feature-flags", "enabled zoom integration connections")
		} else {
			insertAuditLogAs(ctx, user, "feature-flags", "disabled zoom integration connections")
		}
	}
	if !slices.Equal(prevZoomRoles, zoomRoles) {
		insertAuditLogAs(ctx, user, "feature-flags", "updated zoom visible roles: "+formatVisibleRolesAudit(zoomRoles))
	}
	if prevGoogleEnabled != googleEnabled {
		if googleEnabled {
			insertAuditLogAs(ctx, user, "feature-flags", "enabled google calendar integration connections")
		} else {
			insertAuditLogAs(ctx, user, "feature-flags", "disabled google calendar integration connections")
		}
	}
	if !slices.Equal(prevGoogleRoles, googleRoles) {
		insertAuditLogAs(ctx, user, "feature-flags", "updated google calendar visible roles: "+formatVisibleRolesAudit(googleRoles))
	}
	if prevGracePeriod != gracePeriod {
		insertAuditLogAs(ctx, user, "feature-flags", fmt.Sprintf("updated class overdue grace period to %d minutes", gracePeriod))
	}

	setSuccessFlash(w, "Feature flags saved successfully")
	HttpRedirect(w, r, "/feature-flags")
}

func parseClassOverdueGracePeriodFromForm(r *http.Request) (int64, error) {
	raw := strings.TrimSpace(r.FormValue("class_overdue_grace_period_minutes"))
	if raw == "" {
		return constants.DefaultClassOverdueGracePeriodMinutes, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Class overdue grace period must be a whole number of minutes")
	}
	if value < 0 {
		return 0, fmt.Errorf("Class overdue grace period cannot be negative")
	}
	if value > constants.MaxClassOverdueGracePeriodMinutes {
		return 0, fmt.Errorf("Class overdue grace period cannot exceed %d minutes", constants.MaxClassOverdueGracePeriodMinutes)
	}
	return value, nil
}
