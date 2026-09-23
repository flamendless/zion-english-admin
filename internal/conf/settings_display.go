package conf

import (
	"fmt"
	"strconv"
	"strings"

	"zion-english/internal/constants"
	"zion-english/internal/teacherintrovideo"
)

type SensitiveFieldKey string

const (
	SensitiveFieldSecret                 SensitiveFieldKey = "secret"
	SensitiveFieldSuperuserPassword      SensitiveFieldKey = "superuser_password"
	SensitiveFieldZoomClientSecret       SensitiveFieldKey = "zoom_client_secret"
	SensitiveFieldGoogleCalendarSecret   SensitiveFieldKey = "google_calendar_client_secret"
	SensitiveFieldR2AccessKeyID          SensitiveFieldKey = "r2_access_key_id"
	SensitiveFieldR2SecretAccessKey      SensitiveFieldKey = "r2_secret_access_key"
	SensitiveFieldCloudflareAPIToken     SensitiveFieldKey = "cloudflare_api_token"
)

type SettingsField struct {
	Label     string
	EnvKey    string
	Value     string
	Monospace bool
	Sensitive bool
	FieldKey  SensitiveFieldKey
}

type SettingsSection struct {
	Title  string
	Fields []SettingsField
}

type SettingsRuntime struct {
	VersionSummary           string
	PublicURL                string
	ListenAddress            string
	HTTPS                    bool
	BasePath                 string
	TLSCertPath              string
	DatabasePath             string
	ZoomConfigured           bool
	GoogleCalendarConfigured bool
	MeetingService           string
}

func BuildSettingsSections(cfg *Config, runtime SettingsRuntime) []SettingsSection {
	return []SettingsSection{
		buildConfigurationSection(cfg, runtime),
		buildIntegrationsSection(cfg, runtime),
		buildEnvironmentVarsSection(cfg),
	}
}

func ParseSensitiveFieldKey(raw string) (SensitiveFieldKey, error) {
	switch SensitiveFieldKey(raw) {
	case SensitiveFieldSecret,
		SensitiveFieldSuperuserPassword,
		SensitiveFieldZoomClientSecret,
		SensitiveFieldGoogleCalendarSecret,
		SensitiveFieldR2AccessKeyID,
		SensitiveFieldR2SecretAccessKey,
		SensitiveFieldCloudflareAPIToken:
		return SensitiveFieldKey(raw), nil
	default:
		return "", ErrSettingsInvalidField
	}
}

func ResolveSensitiveValue(key SensitiveFieldKey, cfg *Config) string {
	switch key {
	case SensitiveFieldSecret:
		return cfg.Secret
	case SensitiveFieldSuperuserPassword:
		return cfg.SuperuserPassword
	case SensitiveFieldZoomClientSecret:
		return cfg.Meeting.Zoom.ClientSecret
	case SensitiveFieldGoogleCalendarSecret:
		return cfg.Calendar.Google.ClientSecret
	case SensitiveFieldR2AccessKeyID:
		return cfg.Storage.R2.AccessKeyID
	case SensitiveFieldR2SecretAccessKey:
		return cfg.Storage.R2.SecretAccessKey
	case SensitiveFieldCloudflareAPIToken:
		return cfg.Storage.R2.APIToken
	default:
		return ""
	}
}

func buildConfigurationSection(cfg *Config, runtime SettingsRuntime) SettingsSection {
	fields := []SettingsField{
		plainField(constants.SettingsFieldVersion, "", runtime.VersionSummary, false),
		plainField(constants.SettingsFieldEnvironment, "APP_ENV", cfg.AppEnv, false),
		plainField(constants.SettingsFieldPort, "PORT", strconv.Itoa(cfg.Port), false),
		plainField(constants.SettingsFieldPublicURL, "", runtime.PublicURL, true),
		plainField(constants.SettingsFieldListenAddress, "", runtime.ListenAddress, true),
		plainField(constants.SettingsFieldHTTPS, "", strconv.FormatBool(runtime.HTTPS), false),
		plainField(constants.SettingsFieldBasePath, "", settingsValueOrUnset(runtime.BasePath), true),
		plainField(constants.SettingsFieldDatabasePath, "", runtime.DatabasePath, true),
		plainField(constants.SettingsFieldStorageBackend, "", settingsStorageBackendLabel(cfg), false),
		plainField(constants.SettingsFieldStorageBucket, "R2_BUCKET", cfg.Storage.R2.Bucket, false),
		plainField(constants.SettingsFieldSuperuserUsername, "SUPERUSER_USERNAME", cfg.SuperuserUsername, false),
		plainField(constants.SettingsFieldMeetingService, "MEETING_SERVICE", settingsValueOrUnset(cfg.Meeting.Service), false),
	}
	if runtime.HTTPS && runtime.TLSCertPath != "" {
		fields = append(fields, plainField(constants.SettingsFieldTLSCert, "", runtime.TLSCertPath, true))
	}
	return SettingsSection{
		Title:  constants.SettingsSectionConfiguration,
		Fields: fields,
	}
}

func buildIntegrationsSection(cfg *Config, runtime SettingsRuntime) SettingsSection {
	return SettingsSection{
		Title: constants.SettingsSectionIntegrations,
		Fields: []SettingsField{
			plainField(constants.SettingsFieldIntegrationZoom, "", settingsZoomIntegrationLine(cfg, runtime.ZoomConfigured), false),
			plainField(constants.SettingsFieldIntegrationGoogle, "", settingsGoogleCalendarIntegrationLine(cfg, runtime.GoogleCalendarConfigured), false),
			plainField(constants.SettingsFieldIntegrationStorage, "", settingsStorageIntegrationLine(cfg), false),
			plainField(constants.SettingsFieldIntegrationFfmpeg, "", settingsToolIntegrationLine(teacherintrovideo.FfmpegAvailable()), false),
			plainField(constants.SettingsFieldIntegrationFfprobe, "", settingsToolIntegrationLine(teacherintrovideo.FfprobeAvailable()), false),
		},
	}
}

func buildEnvironmentVarsSection(cfg *Config) SettingsSection {
	fields := []SettingsField{
		sensitiveField("Secret", "SECRET", SensitiveFieldSecret, cfg),
		plainField("Superuser username", "SUPERUSER_USERNAME", cfg.SuperuserUsername, false),
		sensitiveField("Superuser password", "SUPERUSER_PASSWORD", SensitiveFieldSuperuserPassword, cfg),
		plainField("Meeting service", "MEETING_SERVICE", cfg.Meeting.Service, false),
		plainField("Zoom client ID", "ZOOM_CLIENT_ID", cfg.Meeting.Zoom.ClientID, true),
		sensitiveField("Zoom client secret", "ZOOM_CLIENT_SECRET", SensitiveFieldZoomClientSecret, cfg),
		plainField("Zoom redirect URI", "ZOOM_REDIRECT_URI", cfg.Meeting.Zoom.RedirectURI, true),
		plainField("Zoom authorize URL", "ZOOM_AUTHORIZE_URL", cfg.Meeting.Zoom.AuthorizeURL, true),
		plainField("Google Calendar client ID", "GOOGLE_CALENDAR_CLIENT_ID", cfg.Calendar.Google.ClientID, true),
		sensitiveField("Google Calendar client secret", "GOOGLE_CALENDAR_CLIENT_SECRET", SensitiveFieldGoogleCalendarSecret, cfg),
		plainField("Google Calendar redirect URI", "GOOGLE_CALENDAR_REDIRECT_URI", cfg.Calendar.Google.RedirectURI, true),
		plainField("R2 account ID", "R2_ACCOUNT_ID", cfg.Storage.R2.AccountID, true),
		plainField("R2 bucket", "R2_BUCKET", cfg.Storage.R2.Bucket, false),
		sensitiveField("R2 access key ID", "R2_ACCESS_KEY_ID", SensitiveFieldR2AccessKeyID, cfg),
		sensitiveField("R2 secret access key", "R2_SECRET_ACCESS_KEY", SensitiveFieldR2SecretAccessKey, cfg),
		sensitiveField("Cloudflare API token", "CLOUDFLARE_API_TOKEN", SensitiveFieldCloudflareAPIToken, cfg),
	}
	return SettingsSection{
		Title:  constants.SettingsSectionEnvironmentVars,
		Fields: fields,
	}
}

func plainField(label, envKey, value string, monospace bool) SettingsField {
	return SettingsField{
		Label:     label,
		EnvKey:    envKey,
		Value:     settingsValueOrUnset(value),
		Monospace: monospace,
	}
}

func sensitiveField(label, envKey string, key SensitiveFieldKey, cfg *Config) SettingsField {
	return SettingsField{
		Label:     label,
		EnvKey:    envKey,
		Value:     MaskSecret(ResolveSensitiveValue(key, cfg)),
		Monospace: true,
		Sensitive: true,
		FieldKey:  key,
	}
}

func settingsValueOrUnset(value string) string {
	if strings.TrimSpace(value) == "" {
		return constants.StartupValueUnset
	}
	return value
}

func settingsStorageBackendLabel(cfg *Config) string {
	if cfg.R2Enabled() {
		return "r2"
	}
	return "local"
}

func settingsZoomIntegrationLine(cfg *Config, configured bool) string {
	if configured {
		return "configured"
	}
	missing := settingsMissingZoomFields(cfg)
	if len(missing) == 0 {
		return "not configured"
	}
	return fmt.Sprintf("not configured (missing: %s)", strings.Join(missing, ", "))
}

func settingsGoogleCalendarIntegrationLine(cfg *Config, configured bool) string {
	if configured {
		return "configured"
	}
	missing := settingsMissingGoogleCalendarFields(cfg)
	if len(missing) == 0 {
		return "not configured"
	}
	return fmt.Sprintf("not configured (missing: %s)", strings.Join(missing, ", "))
}

func settingsStorageIntegrationLine(cfg *Config) string {
	if cfg.R2Enabled() {
		return fmt.Sprintf("r2 (bucket=%s)", settingsValueOrUnset(cfg.Storage.R2.Bucket))
	}
	return "local (data/ and tmp/)"
}

func settingsToolIntegrationLine(available bool) string {
	if available {
		return "configured"
	}
	return "not configured (not found in PATH)"
}

func settingsMissingZoomFields(cfg *Config) []string {
	zoom := cfg.Meeting.Zoom
	missing := make([]string, 0, 4)
	if zoom.ClientID == "" {
		missing = append(missing, "ZOOM_CLIENT_ID")
	}
	if zoom.ClientSecret == "" {
		missing = append(missing, "ZOOM_CLIENT_SECRET")
	}
	if zoom.RedirectURI == "" {
		missing = append(missing, "ZOOM_REDIRECT_URI")
	}
	if zoom.AuthorizeURL == "" {
		missing = append(missing, "ZOOM_AUTHORIZE_URL")
	}
	return missing
}

func settingsMissingGoogleCalendarFields(cfg *Config) []string {
	gcal := cfg.Calendar.Google
	missing := make([]string, 0, 3)
	if gcal.ClientID == "" {
		missing = append(missing, "GOOGLE_CALENDAR_CLIENT_ID")
	}
	if gcal.ClientSecret == "" {
		missing = append(missing, "GOOGLE_CALENDAR_CLIENT_SECRET")
	}
	if gcal.RedirectURI == "" {
		missing = append(missing, "GOOGLE_CALENDAR_REDIRECT_URI")
	}
	return missing
}
