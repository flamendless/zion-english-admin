package startup

import (
	"strings"
	"testing"

	"zion-english/internal/conf"
	"zion-english/internal/version"
)

func TestVisibleLenIgnoresANSI(t *testing.T) {
	if got := visibleLen(ansiBold + "test" + ansiReset); got != 4 {
		t.Fatalf("got %d want 4", got)
	}
}

func TestStartupBannerIncludesVersionAndEnv(t *testing.T) {
	oldCommit := version.Commit
	oldTag := version.Tag
	t.Cleanup(func() {
		version.Commit = oldCommit
		version.Tag = oldTag
	})

	version.Commit = "testcommit123"
	version.Tag = "v1.2.3"

	lines := startupBannerLines(testOptions())
	body := strings.Join(lines, "\n")
	if !strings.Contains(body, "tag=v1.2.3  commit=testcommit123") {
		t.Fatalf("missing version line: %s", body)
	}
	if !strings.Contains(body, "(local)") {
		t.Fatalf("missing env label: %s", body)
	}
}

func TestStartupBannerMasksSensitiveValues(t *testing.T) {
	opts := testOptions()
	opts.Cfg.Secret = "change-me-dev-only-secret!!"

	lines := startupBannerLines(opts)
	body := strings.Join(lines, "\n")

	if strings.Contains(body, "change-me-dev-only-secret!!") {
		t.Fatal("raw secret leaked into banner")
	}
	if !strings.Contains(body, conf.MaskSecret("change-me-dev-only-secret!!")) {
		t.Fatalf("expected masked secret in banner: %s", body)
	}
}

func TestStartupIntegrationLines(t *testing.T) {
	opts := testOptions()
	opts.Integrations = IntegrationStatus{
		ZoomConfigured:           true,
		GoogleCalendarConfigured: false,
		MeetingService:           "zoom",
	}

	lines := startupIntegrationLines(opts)
	body := strings.Join(lines, "\n")
	if !strings.Contains(body, "Zoom: configured") {
		t.Fatalf("expected zoom configured: %s", body)
	}
	if !strings.Contains(body, "Google Calendar: not configured") {
		t.Fatalf("expected google calendar not configured: %s", body)
	}
	if !strings.Contains(body, "GOOGLE_CALENDAR_CLIENT_ID") {
		t.Fatalf("expected missing google calendar fields: %s", body)
	}
}

func TestListenURLLocal(t *testing.T) {
	opts := testOptions()
	opts.ListenPort = ":8080"
	opts.BasePath = "/zion-english-admin"
	got := ListenURL(opts)
	want := "http://localhost:8080/zion-english-admin"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func testOptions() Options {
	return Options{
		Cfg: &conf.Config{
			AppEnv:            conf.EnvLocal,
			Port:              8080,
			Secret:            "secret",
			SuperuserUsername: "admin",
		},
		ListenPort: ":8080",
		BasePath:   "/zion-english-admin",
		Integrations: IntegrationStatus{
			MeetingService: "zoom",
		},
	}
}
