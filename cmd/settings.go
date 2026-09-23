package cmd

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/startup"
	"zion-english/internal/version"
)

var serverStartupOpts startup.Options

func handleSettings(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	data := frontend.SettingsData{
		Title:    constants.SettingsPageTitle,
		Subtitle: constants.SettingsPageSubtitle,
		Sections: conf.BuildSettingsSections(serverStartupOpts.Cfg, buildSettingsRuntime(serverStartupOpts)),
	}
	if err := frontend.Settings(data).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSettingsReveal(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		renderSettingsRevealError(w, r, ErrSettingsInvalidPassword.Error())
		return
	}

	fieldKey, err := conf.ParseSensitiveFieldKey(strings.TrimSpace(r.FormValue("field")))
	if err != nil {
		renderSettingsRevealError(w, r, ErrSettingsInvalidField.Error())
		return
	}

	cfg := conf.Conf()
	password := r.FormValue("password")
	if subtle.ConstantTimeCompare([]byte(password), []byte(cfg.SuperuserPassword)) != 1 {
		renderSettingsRevealError(w, r, ErrSettingsInvalidPassword.Error())
		return
	}

	value := conf.ResolveSensitiveValue(fieldKey, cfg)
	user := auth.GetUser(ctx)
	insertAuditLogAs(ctx, user, "settings", fmt.Sprintf("revealed %s", string(fieldKey)))

	w.Header().Set("HX-Trigger", "settings-revealed")
	if err := frontend.SettingsRevealSuccess(string(fieldKey), value).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderSettingsRevealError(w http.ResponseWriter, r *http.Request, message string) {
	if err := frontend.SettingsRevealError(message).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildSettingsRuntime(opts startup.Options) conf.SettingsRuntime {
	runtime := conf.SettingsRuntime{
		VersionSummary:           version.Get().Summary(),
		PublicURL:                startup.ListenURL(opts),
		ListenAddress:            strings.TrimPrefix(opts.ListenPort, ":"),
		HTTPS:                    opts.HTTPS,
		BasePath:                 opts.BasePath,
		DatabasePath:             "data/zion.db",
		ZoomConfigured:           opts.Integrations.ZoomConfigured,
		GoogleCalendarConfigured: opts.Integrations.GoogleCalendarConfigured,
		MeetingService:           opts.Integrations.MeetingService,
	}
	if opts.HTTPS && opts.TLSAddress != "" {
		runtime.TLSCertPath = fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", opts.TLSAddress)
	}
	return runtime
}
