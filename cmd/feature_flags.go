package cmd

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/featureflags"
	"zion-english/internal/utils"
)

func classOverdueGracePeriodMinutes(ctx context.Context) int64 {
	return featureflags.ClassOverdueGracePeriodMinutes(ctx, dbRO)
}

func overdueCutoffPHT(ctx context.Context) string {
	grace := classOverdueGracePeriodMinutes(ctx)
	if grace < 0 {
		grace = 0
	}
	return utils.DateTimePHT(time.Now().Add(-time.Duration(grace) * time.Minute))
}

func handleFeatureFlags(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleFeatureFlagsGet(w, r)
	case http.MethodPost:
		handleFeatureFlagsUpdate(w, r)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleFeatureFlagsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	zoomEnvConfigured := meetingSvc != nil && meetingSvc.IsZoomConfigured()
	googleEnvConfigured := calendarSvc != nil && calendarSvc.IsConfigured()

	zoomEnabled, zoomVisibleRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationZoom)
	googleEnabled, googleVisibleRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar)
	introVideoEnabled, introVideoVisibleRoles, _ := featureflags.GetFlagDefault(ctx, dbRO, constants.FeatureFlagIntroVideoUploads, false)
	introVideoCompressPreset := featureflags.IntroVideoCompressPreset(ctx, dbRO)
	persistentOnboardingEnabled, _, _ := featureflags.GetFlagDefault(ctx, dbRO, constants.FeatureFlagPersistentOnboarding, false)
	roleOptions := constants.AllTeacherRoles()
	introVideoCompressPresetOptions := make([]frontend.FeatureFlagSelectOption, 0, len(constants.IntroVideoCompressPresetOptions()))
	for _, preset := range constants.IntroVideoCompressPresetOptions() {
		introVideoCompressPresetOptions = append(introVideoCompressPresetOptions, frontend.FeatureFlagSelectOption{
			Value: string(preset),
			Label: constants.IntroVideoCompressPresetLabel(preset),
		})
	}

	data := frontend.FeatureFlagsData{
		ClassOverdueGracePeriodMinutes: classOverdueGracePeriodMinutes(ctx),
		PersistentOnboarding: frontend.FeatureFlagBooleanItem{
			Name:          "Persistent onboarding",
			Description:   "Show a sticky onboarding panel for teachers with incomplete setup steps across all pages.",
			Enabled:       persistentOnboardingEnabled,
			FormFieldName: "persistent_onboarding_enabled",
			ToggleLabel:   "Enabled",
		},
		IntroVideoUploads: frontend.FeatureFlagRoleGatedItem{
			Name:          "Intro video uploads",
			Description:   "Allow teachers to upload introduction videos from My Profile. Admins can still review existing uploads when this is off.",
			Enabled:       introVideoEnabled,
			VisibleRoles:  introVideoVisibleRoles,
			RoleOptions:   roleOptions,
			FormFieldName: "intro_video_uploads_enabled",
			FormPrefix:    "intro_video",
		},
		IntroVideoCompressPreset: frontend.FeatureFlagSelectItem{
			Name:          "Intro video compression",
			Description:   "Controls how uploaded intro videos are re-encoded before storage. Original keeps the uploaded file as-is after validation.",
			Options:       introVideoCompressPresetOptions,
			SelectedValue: string(introVideoCompressPreset),
			FormFieldName: "intro_video_compress_preset",
		},
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
	if !requireMethod(w, r, http.MethodPost) {
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
	introVideoEnabled := r.FormValue("intro_video_uploads_enabled") == "on"
	persistentOnboardingEnabled := r.FormValue("persistent_onboarding_enabled") == "on"

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
	introVideoRoles, err := parseFeatureFlagRolesFromForm(r, "intro_video")
	if err != nil {
		setErrorFlash(w, "Select at least one role for intro video upload visibility")
		HttpRedirect(w, r, "/feature-flags")
		return
	}

	prevZoomEnabled, prevZoomRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationZoom)
	prevGoogleEnabled, prevGoogleRoles, _ := featureflags.GetFlag(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar)
	prevIntroVideoEnabled, prevIntroVideoRoles, _ := featureflags.GetFlagDefault(ctx, dbRO, constants.FeatureFlagIntroVideoUploads, false)
	prevIntroVideoCompressPreset := featureflags.IntroVideoCompressPreset(ctx, dbRO)
	prevPersistentOnboardingEnabled, _, _ := featureflags.GetFlagDefault(ctx, dbRO, constants.FeatureFlagPersistentOnboarding, false)
	prevGracePeriod := classOverdueGracePeriodMinutes(ctx)

	gracePeriod, err := parseClassOverdueGracePeriodFromForm(r)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/feature-flags")
		return
	}

	introVideoCompressPreset, err := parseIntroVideoCompressPresetFromForm(r)
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
	if err := featureflags.SetFlag(ctx, dbRW, constants.FeatureFlagIntroVideoUploads, introVideoEnabled, introVideoRoles); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update intro video uploads flag: %v", err))
		HttpRedirect(w, r, "/feature-flags")
		return
	}
	if err := featureflags.SetFlag(ctx, dbRW, constants.FeatureFlagPersistentOnboarding, persistentOnboardingEnabled, featureflags.DefaultVisibleRoles()); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update persistent onboarding flag: %v", err))
		HttpRedirect(w, r, "/feature-flags")
		return
	}
	if err := featureflags.SetIntValue(ctx, dbRW, constants.FeatureFlagClassOverdueGracePeriod, gracePeriod); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update class overdue grace period: %v", err))
		HttpRedirect(w, r, "/feature-flags")
		return
	}
	if err := featureflags.SetStringValue(ctx, dbRW, constants.FeatureFlagIntroVideoCompressPreset, string(introVideoCompressPreset)); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update intro video compression preset: %v", err))
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
	if prevIntroVideoEnabled != introVideoEnabled {
		if introVideoEnabled {
			insertAuditLogAs(ctx, user, "feature-flags", "enabled intro video uploads")
		} else {
			insertAuditLogAs(ctx, user, "feature-flags", "disabled intro video uploads")
		}
	}
	if !slices.Equal(prevIntroVideoRoles, introVideoRoles) {
		insertAuditLogAs(ctx, user, "feature-flags", "updated intro video upload visible roles: "+formatVisibleRolesAudit(introVideoRoles))
	}
	if prevPersistentOnboardingEnabled != persistentOnboardingEnabled {
		if persistentOnboardingEnabled {
			insertAuditLogAs(ctx, user, "feature-flags", "enabled persistent onboarding")
		} else {
			insertAuditLogAs(ctx, user, "feature-flags", "disabled persistent onboarding")
		}
	}
	if prevGracePeriod != gracePeriod {
		insertAuditLogAs(ctx, user, "feature-flags", fmt.Sprintf("updated class overdue grace period to %d minutes", gracePeriod))
	}
	if prevIntroVideoCompressPreset != introVideoCompressPreset {
		insertAuditLogAs(ctx, user, "feature-flags", fmt.Sprintf("updated intro video compression preset to %s", introVideoCompressPreset))
	}

	setSuccessFlash(w, "Feature flags saved successfully")
	HttpRedirect(w, r, "/feature-flags")
}

func parseIntroVideoCompressPresetFromForm(r *http.Request) (constants.IntroVideoCompressPreset, error) {
	raw := strings.TrimSpace(r.FormValue("intro_video_compress_preset"))
	if raw == "" {
		return constants.DefaultIntroVideoCompressPreset(), nil
	}
	if !constants.ValidIntroVideoCompressPreset(raw) {
		return "", fmt.Errorf("Select a valid intro video compression preset")
	}
	return constants.IntroVideoCompressPreset(raw), nil
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
