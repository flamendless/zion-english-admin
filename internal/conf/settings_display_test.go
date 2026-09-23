package conf

import (
	"strings"
	"testing"
)

func TestBuildSettingsSections_masksSecrets(t *testing.T) {
	cfg := &Config{
		AppEnv:            "local",
		Port:              8080,
		SuperuserUsername: "admin",
		SuperuserPassword: "super-secret-password",
		Secret:            "jwt-signing-secret",
		Meeting: MeetingConfig{
			Service: "zoom",
			Zoom: ZoomConfig{
				ClientID:     "zoom-client",
				ClientSecret: "zoom-client-secret-value",
				RedirectURI:  "http://localhost/callback",
				AuthorizeURL: "https://zoom.us/oauth/authorize",
			},
		},
		Calendar: CalendarConfig{
			Google: GoogleCalendarConfig{
				ClientID:     "google-client",
				ClientSecret: "google-client-secret-value",
				RedirectURI:  "http://localhost/gcal/callback",
			},
		},
		Storage: StorageConfig{
			R2: R2Config{
				AccountID:       "acct",
				AccessKeyID:     "r2-access-key",
				SecretAccessKey: "r2-secret-key",
				Bucket:          "bucket",
				APIToken:        "cloudflare-token",
			},
		},
	}
	runtime := SettingsRuntime{
		VersionSummary:           "commit=test",
		PublicURL:                "http://localhost:8080/zion-english-admin",
		ListenAddress:            "8080",
		BasePath:                 "/zion-english-admin",
		DatabasePath:             "data/zion.db",
		ZoomConfigured:           true,
		GoogleCalendarConfigured: true,
		MeetingService:           "zoom",
	}

	sections := BuildSettingsSections(cfg, runtime)
	joined := strings.Join(flattenSectionValues(sections), "\n")

	secrets := []string{
		cfg.Secret,
		cfg.SuperuserPassword,
		cfg.Meeting.Zoom.ClientSecret,
		cfg.Calendar.Google.ClientSecret,
		cfg.Storage.R2.AccessKeyID,
		cfg.Storage.R2.SecretAccessKey,
		cfg.Storage.R2.APIToken,
	}
	for _, secret := range secrets {
		if strings.Contains(joined, secret) {
			t.Fatalf("raw secret leaked in settings output: %q", secret)
		}
	}

	if !strings.Contains(joined, MaskSecret(cfg.Secret)) {
		t.Fatalf("expected masked secret in output")
	}
}

func TestParseSensitiveFieldKey(t *testing.T) {
	key, err := ParseSensitiveFieldKey(string(SensitiveFieldSecret))
	if err != nil || key != SensitiveFieldSecret {
		t.Fatalf("expected secret key, got %q err=%v", key, err)
	}

	_, err = ParseSensitiveFieldKey("not-a-field")
	if err != ErrSettingsInvalidField {
		t.Fatalf("expected ErrSettingsInvalidField, got %v", err)
	}
}

func TestSettingsToolIntegrationLine(t *testing.T) {
	if got := settingsToolIntegrationLine(true); got != "configured" {
		t.Fatalf("available tool line = %q, want configured", got)
	}
	if got := settingsToolIntegrationLine(false); got != "not configured (not found in PATH)" {
		t.Fatalf("missing tool line = %q", got)
	}
}

func TestResolveSensitiveValue(t *testing.T) {
	cfg := &Config{Secret: "abc"}
	if got := ResolveSensitiveValue(SensitiveFieldSecret, cfg); got != "abc" {
		t.Fatalf("got %q", got)
	}
	if got := ResolveSensitiveValue(SensitiveFieldKey("unknown"), cfg); got != "" {
		t.Fatalf("expected empty for unknown key, got %q", got)
	}
}

func flattenSectionValues(sections []SettingsSection) []string {
	var values []string
	for _, section := range sections {
		for _, field := range section.Fields {
			values = append(values, field.Value)
		}
	}
	return values
}
